package backtrace

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/oneclickvirt/backtrace/model"
	. "github.com/oneclickvirt/defaultset"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
)

func newPacketV6(id uint16, _ net.IP, _ int) []byte {
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
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, cm, src, err := conn.ReadFrom(buf)
		if err != nil {
			if model.EnableLoger {
				Logger.Error("读取IPv6响应失败: " + err.Error())
			}
			return err
		}
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("收到IPv6响应: 来源=%v, 跳数=%d", src, cm.HopLimit))
		}
		fromIP := src.(*net.IPAddr).IP
		err = t.serveData(fromIP, buf[:n])
		if err != nil && model.EnableLoger {
			Logger.Warn("处理IPv6数据失败: " + err.Error())
		}
	}
}

// extractIpv6ASNsFromHops 从跃点中提取ASN列表
func extractIpv6ASNsFromHops(hops []*Hop, enableLogger bool) []string {
	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipv6Asn(n.IP.String())
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

// traceIPv6 IPv6追踪函数
func traceIPv6(ch chan Result, i int, offset int) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info(fmt.Sprintf("开始追踪 %s (%s)", model.Ipv6Names[i], model.Ipv6s[i]))
	}
	var allHops [][]*Hop
	var successfulTraces int
	var mu sync.Mutex
	var wg sync.WaitGroup
	// 并发执行3次trace
	for attempt := 1; attempt <= 3; attempt++ {
		wg.Add(1)
		go func(attemptNum int) {
			defer wg.Done()
			if model.EnableLoger {
				Logger.Info(fmt.Sprintf("第%d次尝试追踪 %s (%s)", attemptNum, model.Ipv6Names[i], model.Ipv6s[i]))
			}
			// 先尝试原始IP地址
			hops, err := Trace(net.ParseIP(model.Ipv6s[i]))
			if err != nil {
				if model.EnableLoger {
					Logger.Warn(fmt.Sprintf("第%d次追踪 %s (%s) 失败: %v", attemptNum, model.Ipv6Names[i], model.Ipv6s[i], err))
				}
				// 如果原始IP失败，尝试备选IP
				if tryAltIPs := tryAlternativeIPs(model.Ipv6Names[i], "v6"); len(tryAltIPs) > 0 {
					for _, altIP := range tryAltIPs {
						if model.EnableLoger {
							Logger.Info(fmt.Sprintf("第%d次尝试备选IP %s 追踪 %s", attemptNum, altIP, model.Ipv6Names[i]))
						}
						hops, err = Trace(net.ParseIP(altIP))
						if err == nil && len(hops) > 0 {
							break // 成功找到可用IP
						}
					}
				}
			}
			if err == nil && len(hops) > 0 {
				mu.Lock()
				allHops = append(allHops, hops)
				successfulTraces++
				mu.Unlock()
				if model.EnableLoger {
					Logger.Info(fmt.Sprintf("第%d次追踪 %s (%s) 成功，获得%d个hop", attemptNum, model.Ipv6Names[i], model.Ipv6s[i], len(hops)))
				}
			}
		}(attempt)
	}
	// 等待所有goroutine完成
	wg.Wait()
	// 如果3次都失败
	if successfulTraces == 0 {
		s := fmt.Sprintf("%v %-24s %v", model.Ipv6Names[i], model.Ipv6s[i], Red("检测不到回程路由节点的IP地址"))
		if model.EnableLoger {
			Logger.Warn(fmt.Sprintf("%s (%s) 3次尝试都失败，检测不到回程路由节点的IP地址", model.Ipv6Names[i], model.Ipv6s[i]))
		}
		ch <- Result{i + offset, s}
		return
	}
	// 合并hops结果
	mergedHops := mergeHops(allHops)
	if model.EnableLoger {
		Logger.Info(fmt.Sprintf("%s (%s) 完成%d次成功追踪，合并后获得%d个hop", model.Ipv6Names[i], model.Ipv6s[i], successfulTraces, len(mergedHops)))
	}
	// 从合并后的hops提取ASN
	asns := extractIpv6ASNsFromHops(mergedHops, model.EnableLoger)
	// 处理不同线路
	if len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", model.Ipv6Names[i])
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
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GT", model.Ipv6Names[i], model.Ipv6s[i]))
			}
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append([]string{"AS4809a"}, asns...)
			if model.EnableLoger {
				Logger.Info(fmt.Sprintf("%s (%s) 线路识别为: CN2GIA", model.Ipv6Names[i], model.Ipv6s[i]))
			}
		}
		tempText += fmt.Sprintf("%-24s ", model.Ipv6s[i])
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
		if tempText == (fmt.Sprintf("%v ", model.Ipv6Names[i]) + fmt.Sprintf("%-40s ", model.Ipv6s[i])) {
			tempText += fmt.Sprintf("%v", Red("检测不到已知线路的ASN"))
			if model.EnableLoger {
				Logger.Warn(fmt.Sprintf("%s (%s) 检测不到已知线路的ASN", model.Ipv6Names[i], model.Ipv6s[i]))
			}
		}
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("%s (%s) 追踪完成，最终结果: %s", model.Ipv6Names[i], model.Ipv6s[i], tempText))
		}
		ch <- Result{i + offset, tempText}
	} else {
		s := fmt.Sprintf("%v %-24s %v", model.Ipv6Names[i], model.Ipv6s[i], Red("检测不到回程路由节点的IPV6地址"))
		if model.EnableLoger {
			Logger.Warn(fmt.Sprintf("%s (%s) 检测不到回程路由节点的IPV6地址", model.Ipv6Names[i], model.Ipv6s[i]))
		}
		ch <- Result{i + offset, s}
	}
}
