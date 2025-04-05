package backtrace

import (
	"encoding/binary"
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
)

func newPacketV6(id uint16, dst net.IP, ttl int) []byte {
	// 创建ICMP消息（回显请求）
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Body: &icmp.Echo{
			ID:  int(id),
			Seq: int(id),
		},
	}
	// 序列化ICMP消息
	icmpData, _ := msg.Marshal(nil)
	// 手动创建原始IPv6数据包头部
	ipHeaderBytes := make([]byte, ipv6.HeaderLen)
	// 设置版本和流量类别（第一个字节）
	ipHeaderBytes[0] = (ipv6.Version << 4)
	// 设置下一个头部（协议）
	ipHeaderBytes[6] = ProtocolIPv6ICMP
	// 设置跳数限制
	ipHeaderBytes[7] = byte(ttl)
	// 设置有效载荷长度（2字节字段）
	binary.BigEndian.PutUint16(ipHeaderBytes[4:6], uint16(len(icmpData)))
	// 设置目标地址（最后16个字节）
	copy(ipHeaderBytes[24:40], dst.To16())
	// 合并头部和ICMP数据
	return append(ipHeaderBytes, icmpData...)
}
