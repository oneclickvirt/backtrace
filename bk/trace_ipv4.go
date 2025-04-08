package backtrace

import (
	"fmt"
	"net"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func newPacketV4(id uint16, dst net.IP, ttl int) []byte {
	// TODO: reuse buffers...
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:  int(id),
			Seq: int(id),
		},
	}
	p, _ := msg.Marshal(nil)
	ip := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + len(p),
		TOS:      16,
		ID:       int(id),
		Dst:      dst,
		Protocol: ProtocolICMP,
		TTL:      ttl,
	}
	buf, err := ip.Marshal()
	if err != nil {
		return nil
	}
	return append(buf, p...)
}

// trace IPv4追踪函数
func trace(ch chan Result, i int) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info(fmt.Sprintf("开始追踪 %s (%s)", ipv4Names[i], ipv4s[i]))
	}
	hops, err := Trace(net.ParseIP(ipv4s[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", ipv4Names[i], ipv4s[i], err)

		if EnableLoger {
			Logger.Error(fmt.Sprintf("追踪 %s (%s) 失败: %v", ipv4Names[i], ipv4s[i], err))
		}
		ch <- Result{i, s}
		return
	}
	// 记录每个hop的信息
	if EnableLoger {
		for hopNum, hop := range hops {
			for nodeNum, node := range hop.Nodes {
				Logger.Info(fmt.Sprintf("追踪 %s (%s) - Hop %d, Node %d: %s (RTT: %v)",
					ipv4Names[i], ipv4s[i], hopNum+1, nodeNum+1, node.IP.String(), node.RTT))
			}
		}
	}
	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipAsn(n.IP.String())
			if asn != "" {
				asns = append(asns, asn)
				if EnableLoger {
					Logger.Info(fmt.Sprintf("IP %s 对应的ASN: %s", n.IP.String(), asn))
				}
			}
		}
	}
	// 处理CN2不同路线的区别
	if asns != nil && len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", ipv4Names[i])
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
			if EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GT", ipv4Names[i], ipv4s[i]))
			}
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append([]string{"AS4809a"}, asns...)
			if EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GIA", ipv4Names[i], ipv4s[i]))
			}
		}
		tempText += fmt.Sprintf("%-15s ", ipv4s[i])
		for _, asn := range asns {
			asnDescription := m[asn]
			switch asn {
			case "":
				continue
			case "AS4809": // 被 AS4809a 和 AS4809b 替代了
				continue
			case "AS9929":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + "         "
				}
			case "AS4809a":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + "         "
				}
			case "AS23764":
				if !strings.Contains(tempText, asnDescription) {
					tempText += DarkGreen(asnDescription) + "         "
				}
			case "AS4809b":
				if !strings.Contains(tempText, asnDescription) {
					tempText += Green(asnDescription) + "         "
				}
			case "AS58807":
				if !strings.Contains(tempText, asnDescription) {
					tempText += Green(asnDescription) + "         "
				}
			default:
				if !strings.Contains(tempText, asnDescription) {
					tempText += White(asnDescription) + "         "
				}
			}
		}
		if tempText == (fmt.Sprintf("%v ", ipv4Names[i]) + fmt.Sprintf("%-15s ", ipv4s[i])) {
			tempText += fmt.Sprintf("%v", Red("检测不到已知线路的ASN"))

			if EnableLoger {
				Logger.Warn(fmt.Sprintf("%s (%s) 检测不到已知线路的ASN", ipv4Names[i], ipv4s[i]))
			}
		}
		if EnableLoger {
			Logger.Info(fmt.Sprintf("%s (%s) 追踪完成，结果: %s", ipv4Names[i], ipv4s[i], tempText))
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", ipv4Names[i], ipv4s[i], Red("检测不到回程路由节点的IP地址"))

		if EnableLoger {
			Logger.Warn(fmt.Sprintf("%s (%s) 检测不到回程路由节点的IP地址", ipv4Names[i], ipv4s[i]))
		}
		ch <- Result{i, s}
	}
}
