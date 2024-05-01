//go:build linux
// +build linux

package backtrace

import (
	"net"
	"syscall"
)

func (t *Tracer) listen(network string, laddr *net.IPAddr) (*net.IPConn, error) {
	conn, err := net.ListenIP(network, laddr)
	if err != nil {
		return nil, err
	}
	raw, err := conn.SyscallConn()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_ = raw.Control(func(fd uintptr) {
		err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
