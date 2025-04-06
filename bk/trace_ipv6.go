package backtrace

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	. "github.com/oneclickvirt/defaultset"
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

func (t *Tracer) serveIPv6(conn *ipv6.PacketConn) error {
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, _, from, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}
		fromIP := from.(*net.IPAddr).IP
		err = t.serveData(fromIP, buf[:n])
		if err != nil {
			continue
		}
	}
}

// IPv6追踪函数
func traceIPv6(ch chan Result, i int, offset int) {
	hops, err := Trace(net.ParseIP(ipv6s[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-40s %v", ipv6Names[i], ipv6s[i], err)
		ch <- Result{i + offset, s} // 注意这里加了offset
		return
	}

	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipAsn(n.IP.String())
			if asn != "" {
				asns = append(asns, asn)
			}
		}
	}

	// 处理路由信息
	if asns != nil && len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", ipv6Names[i])
		hasAS4134 := false
		hasAS4809 := false
		for _, asn := range asns {
			if asn == "AS4134" {
				hasAS4134 = true
			}
			if asn == "AS4809" {
				hasAS4809 = true
			}
		}
		// 判断是否包含 AS4134 和 AS4809
		if hasAS4134 && hasAS4809 {
			// 同时包含 AS4134 和 AS4809 属于 CN2GT
			asns = append([]string{"AS4809b"}, asns...)
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append([]string{"AS4809a"}, asns...)
		}
		tempText += fmt.Sprintf("%-40s ", ipv6s[i])
		for _, asn := range asns {
			asnDescription := m[asn]
			switch asn {
			case "":
				continue
			case "AS4809": // 被 AS4809a 和 AS4809b 替代了
				continue
			case "AS9929":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + " "
				}
			case "AS4809a":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + " "
				}
			case "AS23764":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + " "
				}
			case "AS4809b":
				if !strings.Contains(tempText, asnDescription) {
					tempText += Green(asnDescription) + " "
				}
			case "AS58807":
				if !strings.Contains(tempText, asnDescription) {
					tempText += Green(asnDescription) + " "
				}
			default:
				if !strings.Contains(tempText, asnDescription) {
					tempText += White(asnDescription) + " "
				}
			}
		}
		if tempText == (fmt.Sprintf("%v ", ipv6Names[i]) + fmt.Sprintf("%-40s ", ipv6s[i])) {
			tempText += fmt.Sprintf("%v", Red("检测不到已知线路的ASN"))
		}
		ch <- Result{i + offset, tempText}
	} else {
		s := fmt.Sprintf("%v %-40s %v", ipv6Names[i], ipv6s[i], Red("检测不到回程路由节点的IP地址"))
		ch <- Result{i + offset, s}
	}
}
