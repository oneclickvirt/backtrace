package backtrace

import (
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
)

func newPacketV6(id uint16, dst net.IP, ttl int) ([]byte, error) {
	// 创建ICMPv6 Echo请求消息
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  int(id),
			Data: []byte("TRACEROUTE"),
		},
	}

	// 直接序列化ICMPv6消息
	// 第一个参数是协议号，对于ICMPv6应该是58
	return msg.Marshal(nil)
}

func (t *Tracer) sendRequestV6(dst net.IP, ttl int) (*packet, error) {
	id := uint16(atomic.AddUint32(&t.seq, 1))
	// 创建ICMPv6消息
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  int(id),
			Data: []byte("TRACEROUTE"),
		},
	}
	// 序列化ICMPv6消息
	b, err := msg.Marshal(nil)
	if err != nil {
		return nil, err
	}
	// 获取底层连接
	ipv6Conn, err := t.getIPv6Conn()
	if err != nil {
		return nil, err
	}
	// 设置IPv6数据包的跳数限制(TTL)
	if err := ipv6Conn.SetHopLimit(ttl); err != nil {
		return nil, err
	}
	// 发送数据包
	if _, err := ipv6Conn.WriteTo(b, nil, &net.IPAddr{IP: dst}); err != nil {
		return nil, err
	}
	// 创建一个数据包记录，用于后续匹配回复
	req := &packet{dst, id, ttl, time.Now()}
	return req, nil
}

// getIPv6Conn 获取IPv6的PacketConn接口
func (t *Tracer) getIPv6Conn() (*ipv6.PacketConn, error) {
	if t.ipv6conn != nil {
		return t.ipv6conn, nil
	}
	// 创建一个UDP连接
	c, err := net.ListenPacket("udp6", "::")
	if err != nil {
		return nil, err
	}
	// 将其包装为IPv6 PacketConn
	p := ipv6.NewPacketConn(c)
	// 设置控制消息
	if err := p.SetControlMessage(ipv6.FlagHopLimit|ipv6.FlagSrc|ipv6.FlagDst|ipv6.FlagInterface, true); err != nil {
		c.Close()
		return nil, err
	}
	t.ipv6conn = p
	return p, nil
}

func (t *Tracer) serveIPv6(conn *ipv6.PacketConn) error {
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, cm, src, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}
		// 从控制消息中获取跳数限制
		hopLimit := 0
		if cm != nil {
			hopLimit = cm.HopLimit
		}
		// 解析ICMP消息
		msg, err := icmp.ParseMessage(ProtocolIPv6ICMP, buf[:n])
		if err != nil {
			continue
		}
		// 根据消息类型处理
		switch msg.Type {
		case ipv6.ICMPTypeEchoReply:
			echo := msg.Body.(*icmp.Echo)
			t.serveReply(src.(*net.IPAddr).IP, &packet{src.(*net.IPAddr).IP, uint16(echo.ID), hopLimit, time.Now()})
		case ipv6.ICMPTypeTimeExceeded:
			// 处理超时消息，获取原始数据包
			data := msg.Body.(*icmp.TimeExceeded).Data
			// 尝试提取嵌入的原始Echo请求
			if len(data) < 8 { // 至少需要IPv6头部前8个字节
				continue
			}
			// 跳过IPv6头部和ICMPv6头部，简化处理，实际可能需要更复杂的解析
			innerMsg, err := icmp.ParseMessage(ProtocolIPv6ICMP, data[48:])
			if err != nil {
				continue
			}
			if echo, ok := innerMsg.Body.(*icmp.Echo); ok {
				t.serveReply(src.(*net.IPAddr).IP, &packet{src.(*net.IPAddr).IP, uint16(echo.ID), hopLimit, time.Now()})
			}
		}
	}
}
