package elevate

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteExW = shell32.NewProc("ShellExecuteExW")
)

const (
	SEE_MASK_NOCLOSEPROCESS = 0x00000040
	SW_HIDE                 = 0
	ERROR_CANCELLED         = 1223
)

type SHELLEXECUTEINFO struct {
	cbSize         uint32
	fMask          uint32
	hwnd           uintptr
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       uintptr
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      uintptr
	dwHotKey       uint32
	hIconOrMonitor uintptr
	hProcess       windows.Handle
}

var ErrCancelled = errors.New("elevation cancelled by user")

// IsElevated checks if the current process is running with admin privileges
func IsElevated() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()
	return token.IsElevated()
}

// RunAsAdminWait executes a command with UAC elevation and waits for completion
func RunAsAdminWait(file string, args string) (uint32, error) {
	var sei SHELLEXECUTEINFO
	sei.cbSize = uint32(unsafe.Sizeof(sei))
	sei.fMask = SEE_MASK_NOCLOSEPROCESS
	sei.lpVerb = syscall.StringToUTF16Ptr("runas")
	sei.lpFile = syscall.StringToUTF16Ptr(file)
	sei.nShow = SW_HIDE

	if args != "" {
		sei.lpParameters = syscall.StringToUTF16Ptr(args)
	}

	ret, _, err := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&sei)))
	if ret == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno == ERROR_CANCELLED {
			return 0, ErrCancelled
		}
		return 0, fmt.Errorf("ShellExecuteEx failed: %w", err)
	}

	if sei.hProcess == 0 {
		return 0, errors.New("no process handle returned")
	}
	defer windows.CloseHandle(sei.hProcess)

	_, err = windows.WaitForSingleObject(sei.hProcess, windows.INFINITE)
	if err != nil {
		return 0, err
	}

	var exitCode uint32
	if err := windows.GetExitCodeProcess(sei.hProcess, &exitCode); err != nil {
		return 0, err
	}
	return exitCode, nil
}
