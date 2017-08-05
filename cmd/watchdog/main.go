package main

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	watchdogDevice  string = "/dev/watchdog"
	ioctlSetTimeout uint   = 0xc0045706
	ioctlKeepAlive  uint   = 0x80045705
)

func ioctl(f *os.File, cmd uint, valptr int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(cmd), uintptr(unsafe.Pointer(&valptr)))
	if err != 0 {
		return err
	}
	return nil
}

func main() {
	f, err := os.Open(watchdogDevice)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := ioctl(f, ioctlSetTimeout, 120); err != nil {
		panic(err)
	}

	for {
		if err := ioctl(f, ioctlKeepAlive, 0); err != nil {
			panic(err)
		}
		time.Sleep(30 * time.Second)
	}

}
