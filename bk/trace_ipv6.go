package bk

import (
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
)

func newPacketV6(id uint16, dst net.IP, ttl int) []byte {
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Body: &icmp.Echo{
			ID:  int(id),
			Seq: int(id),
		},
	}
	p, _ := msg.Marshal(nil)
	ip := &ipv6.Header{
		Version:    ipv6.Version,
		NextHeader: ProtocolIPv6ICMP,
		HopLimit:   ttl,
		Dst:        dst,
	}
	buf, err := ip.Marshal()
	if err != nil {
		return nil
	}
	return append(buf, p...)
}
