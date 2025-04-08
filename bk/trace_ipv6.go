package backtrace

import (
	"fmt"
	"net"
	"strings"

	. "github.com/oneclickvirt/defaultset"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
)

func newPacketV6(id uint16, dst net.IP, ttl int) []byte {
	// 使用ipv6包的Echo请求
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  int(id),
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	// 序列化ICMP消息
	icmpBytes, _ := msg.Marshal(nil)
	return icmpBytes
}

func (t *Tracer) serveIPv6(conn *ipv6.PacketConn) error {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, cm, src, err := conn.ReadFrom(buf)
		if err != nil {
			if EnableLoger {
				Logger.Error("读取IPv6响应失败: " + err.Error())
			}
			return err
		}
		if EnableLoger {
			Logger.Info(fmt.Sprintf("收到IPv6响应: 来源=%v, 跳数=%d", src, cm.HopLimit))
		}
		fromIP := src.(*net.IPAddr).IP
		err = t.serveData(fromIP, buf[:n])
		if err != nil && EnableLoger {
			Logger.Warn("处理IPv6数据失败: " + err.Error())
		}
	}
}

// traceIPv6 IPv6追踪函数
func traceIPv6(ch chan Result, i int, offset int) {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info(fmt.Sprintf("开始追踪 %s (%s)", ipv6Names[i], ipv6s[i]))
	}
	hops, err := Trace(net.ParseIP(ipv6s[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-24s %v", ipv6Names[i], ipv6s[i], err)
		if EnableLoger {
			Logger.Error(fmt.Sprintf("追踪 %s (%s) 失败: %v", ipv6Names[i], ipv6s[i], err))
		}
		ch <- Result{i + offset, s}
		return
	}
	// 记录每个hop的信息
	if EnableLoger {
		for hopNum, hop := range hops {
			for nodeNum, node := range hop.Nodes {
				Logger.Info(fmt.Sprintf("追踪 %s (%s) - Hop %d, Node %d: %s (RTT: %v)",
					ipv6Names[i], ipv6s[i], hopNum+1, nodeNum+1, node.IP.String(), node.RTT))
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
			if EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GT", ipv6Names[i], ipv6s[i]))
			}
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append([]string{"AS4809a"}, asns...)
			if EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GIA", ipv6Names[i], ipv6s[i]))
			}
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

			if EnableLoger {
				Logger.Warn(fmt.Sprintf("%s (%s) 检测不到已知线路的ASN", ipv6Names[i], ipv6s[i]))
			}
		}
		if EnableLoger {
			Logger.Info(fmt.Sprintf("%s (%s) 追踪完成，结果: %s", ipv6Names[i], ipv6s[i], tempText))
		}
		ch <- Result{i + offset, tempText}
	} else {
		s := fmt.Sprintf("%v %-40s %v", ipv6Names[i], ipv6s[i], Red("检测不到回程路由节点的IP地址"))
		if EnableLoger {
			Logger.Warn(fmt.Sprintf("%s (%s) 检测不到回程路由节点的IP地址", ipv6Names[i], ipv6s[i]))
		}
		ch <- Result{i + offset, s}
	}
}
