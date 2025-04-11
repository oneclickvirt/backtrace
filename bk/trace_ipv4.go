package backtrace

import (
	"fmt"
	"net"
	"strings"

	"github.com/oneclickvirt/backtrace/model"
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

// extractIpv4ASNsFromHops 从跃点中提取ASN列表
func extractIpv4ASNsFromHops(hops []*Hop, enableLogger bool) []string {
	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipv4Asn(n.IP.String())
			if asn != "" {
				asns = append(asns, asn)
				if enableLogger {
					Logger.Info(fmt.Sprintf("IP %s 对应的ASN: %s", n.IP.String(), asn))
				}
			}
		}
	}
	return asns
}

// trace IPv4追踪函数
func trace(ch chan Result, i int) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info(fmt.Sprintf("开始追踪 %s (%s)", model.Ipv4Names[i], model.Ipv4s[i]))
	}
	// 先尝试原始IP地址
	hops, err := Trace(net.ParseIP(model.Ipv4s[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", model.Ipv4Names[i], model.Ipv4s[i], Red("检测不到回程路由节点的IP地址"))
		if model.EnableLoger {
			Logger.Error(fmt.Sprintf("追踪 %s (%s) 失败: %v", model.Ipv4Names[i], model.Ipv4s[i], err))
		}
		ch <- Result{i, s}
	}
	asns := extractIpv4ASNsFromHops(hops, model.EnableLoger)
	// 如果没有找到ASN，尝试备选IP
	if len(asns) == 0 {
		// 尝试从IcmpTargets获取备选IP
		if tryAltIPs := tryAlternativeIPs(model.Ipv4Names[i], "v4"); len(tryAltIPs) > 0 {
			for _, altIP := range tryAltIPs {
				if model.EnableLoger {
					Logger.Info(fmt.Sprintf("尝试备选IP %s 追踪 %s", altIP, model.Ipv4Names[i]))
				}
				hops, err = Trace(net.ParseIP(altIP))
				if err == nil && len(hops) > 0 {
					newAsns := extractIpv4ASNsFromHops(hops, model.EnableLoger)
					asns = append(asns, newAsns...)
					if len(newAsns) > 0 {
						break // 成功找到可用IP
					}
				}
			}
		}
	}
	asns = removeDuplicates(asns)
	// // 记录每个hop的信息
	// if model.EnableLoger {
	// 	for hopNum, hop := range hops {
	// 		for nodeNum, node := range hop.Nodes {
	// 			Logger.Info(fmt.Sprintf("追踪 %s (%s) - Hop %d, Node %d: %s (RTT: %v)",
	// 				model.Ipv4Names[i], model.Ipv4s[i], hopNum+1, nodeNum+1, node.IP.String(), node.RTT))
	// 		}
	// 	}
	// }
	// 处理不同线路
	if len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", model.Ipv4Names[i])
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
			if model.EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GT", model.Ipv4Names[i], model.Ipv4s[i]))
			}
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append([]string{"AS4809a"}, asns...)
			if model.EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GIA", model.Ipv4Names[i], model.Ipv4s[i]))
			}
		}
		tempText += fmt.Sprintf("%-24s ", model.Ipv4s[i])
		for _, asn := range asns {
			asnDescription := model.M[asn]
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
		if tempText == (fmt.Sprintf("%v ", model.Ipv4Names[i]) + fmt.Sprintf("%-15s ", model.Ipv4s[i])) {
			tempText += fmt.Sprintf("%v", Red("检测不到已知线路的ASN"))

			if model.EnableLoger {
				Logger.Warn(fmt.Sprintf("%s (%s) 检测不到已知线路的ASN", model.Ipv4Names[i], model.Ipv4s[i]))
			}
		}
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("%s (%s) 追踪完成，结果: %s", model.Ipv4Names[i], model.Ipv4s[i], tempText))
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", model.Ipv4Names[i], model.Ipv4s[i], Red("检测不到回程路由节点的IPV4地址"))

		if model.EnableLoger {
			Logger.Warn(fmt.Sprintf("%s (%s) 检测不到回程路由节点的IPV4地址", model.Ipv4Names[i], model.Ipv4s[i]))
		}
		ch <- Result{i, s}
	}
}
