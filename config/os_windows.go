package config

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procGetFolderPath = shell32.NewProc("SHGetFolderPathW")
)

const (
	csidlAppdata = 0x1a
)

func appdataFolderPath() string {
	b := make([]uint16, syscall.MAX_PATH)
	ret, _, err := syscall.Syscall6(procGetFolderPath.Addr(), 5, 0, csidlAppdata, 0, 0, uintptr(unsafe.Pointer(&b[0])), 0)
	if int(ret) != 0 {
		panic(fmt.Sprintf("SHGetFolderPathW : err %d", int(err)))
	}
	return syscall.UTF16ToString(b)
}

// SystemDataDir points to the function that gets the data directory for this system
var SystemDataDir = appdataFolderPath

func firstEnvironmentVariable(vs ...string) string {
	for _, v := range vs {
		val := os.Getenv(v)
		if val != "" {
			return val
		}
	}
	return ""
}

func localHome() string {
	return firstEnvironmentVariable("HOMEPATH", "USERPROFILE")
}
