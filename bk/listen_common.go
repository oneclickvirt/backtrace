//go:build linux || freebsd || openbsd || darwin
// +build linux freebsd openbsd darwin

package backtrace

import (
	"net"
	"syscall"

	. "github.com/oneclickvirt/defaultset"
)

func (t *Tracer) listen(network string, laddr *net.IPAddr) (*net.IPConn, error) {
    conn, err := net.ListenIP(network, laddr)
    if err != nil {
        if EnableLoger {
            Logger.Info(err.Error())
        }
        return nil, err
    }
    raw, err := conn.SyscallConn()
    if err != nil {
        if EnableLoger {
            Logger.Info(err.Error())
        }
        conn.Close()
        return nil, err
    }
    _ = raw.Control(func(fd uintptr) {
        err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
    })
    if err != nil {
        if EnableLoger {
            Logger.Info(err.Error())
        }
        conn.Close()
        return nil, err
    }
    return conn, nil
}