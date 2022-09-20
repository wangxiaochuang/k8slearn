//go:build !windows
// +build !windows

package options

import (
	"syscall"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func permitPortReuse(network, addr string, conn syscall.RawConn) error {
	return conn.Control(func(fd uintptr) {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			klog.Warningf("failed to set SO_REUSEPORT on socket: %v", err)
		}
	})
}

func permitAddressReuse(network, addr string, conn syscall.RawConn) error {
	return conn.Control(func(fd uintptr) {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			klog.Warningf("failed to set SO_REUSEADDR on socket: %v", err)
		}
	})
}
