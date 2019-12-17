package generate

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

// These files get rendered as part of the build process.

// If you add a file to `sliver/` it won't be automatically included
// as part of the build by the server, you have to add it here.

var (
	srcFiles = []string{
		"constants/constants.go",

		"evasion/evasion.go",
		"evasion/evasion_darwin.go",
		"evasion/evasion_linux.go",
		"evasion/evasion_windows.go",
		// C files for DLL
		"dllmain.go",
		"dllmain.h",
		"dllmain.c",

		"handlers/generic-rpc-handlers.go",
		"handlers/generic-tun-handlers.go",
		"handlers/special-handlers.go",
		"handlers/handlers_darwin.go",
		"handlers/handlers_linux.go",
		"handlers/handlers_windows.go",
		"handlers/handlers.go",

		"limits/limits.go",
		"limits/limits_windows.go",
		"limits/limits_darwin.go",
		"limits/limits_linux.go",

		"procdump/dump.go",
		"procdump/dump_windows.go",
		"procdump/dump_linux.go",
		"procdump/dump_darwin.go",

		"proxy/provider_darwin.go",
		"proxy/provider_linux.go",
		"proxy/provider_windows.go",
		"proxy/provider.go",
		"proxy/proxy.go",
		"proxy/url.go",

		"ps/ps.go",
		"ps/ps_windows.go",
		"ps/ps_linux.go",
		"ps/ps_darwin.go",

		"shell/shell.go",
		"shell/shell_windows.go",
		"shell/shell_darwin.go",
		"shell/pty/pty_darwin.go",
		"shell/shell_linux.go",
		"shell/pty/pty_linux.go",

		"shell/pty/run.go",
		"shell/pty/util.go",
		"shell/pty/doc.go",
		"shell/pty/types.go",
		"shell/pty/ztypes_386.go",
		"shell/pty/ztypes_amd64.go",
		"shell/pty/ioctl.go",
		"shell/pty/ioctl_bsd.go",
		"shell/pty/ioctl_darwin.go",
		"shell/pty/pty_unsupported.go",

		"taskrunner/task.go",
		"taskrunner/task_windows.go",
		"taskrunner/task_darwin.go",
		"taskrunner/task_linux.go",

		"priv/priv.go",
		"priv/priv_windows.go",

		"syscalls/syscalls.go",
		"syscalls/syscalls_windows.go",
		"syscalls/types_windows.go",
		"syscalls/zsyscalls_windows.go",

		"transports/crypto.go",
		"transports/tcp-mtls.go",
		"transports/tcp-http.go",
		"transports/udp-dns.go",
		"transports/transports.go",

		"version/version.go",
		"version/version_windows.go",
		"version/version_linux.go",
		"version/version_darwin.go",

		"winhttp/winhttp.go",

		"sliver.go",
	}
)
