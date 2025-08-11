package backtrace

import (
	"strings"
	"time"

	"github.com/oneclickvirt/backtrace/model"
)

func BackTrace(enableIpv6 bool) string {
	if model.CachedIcmpData == "" || model.ParsedIcmpTargets == nil || time.Since(model.CachedIcmpDataFetchTime) > time.Hour {
		model.CachedIcmpData = getData(model.IcmpTargets)
		model.CachedIcmpDataFetchTime = time.Now()
		if model.CachedIcmpData != "" {
			model.ParsedIcmpTargets = parseIcmpTargets(model.CachedIcmpData)
		}
	}
	var builder strings.Builder
	if enableIpv6 {
		ipv4Count := len(model.Ipv4s)
		ipv6Count := len(model.Ipv6s)
		totalCount := ipv4Count + ipv6Count
		var (
			s = make([]string, totalCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range model.Ipv4s {
			go trace(c, i)
		}
		for i := range model.Ipv6s {
			go traceIPv6(c, i, ipv4Count)
		}
	loopIPv4v6:
		for range s {
			select {
			case o := <-c:
				s[o.i] = o.s
			case <-t:
				break loopIPv4v6
			}
		}
		// 收集 IPv4 结果
		for i := 0; i < ipv4Count; i++ {
			if s[i] != "" {
				builder.WriteString(s[i])
				builder.WriteString("\n")
			}
		}
		// 收集 IPv6 结果
		for i := ipv4Count; i < totalCount; i++ {
			if s[i] != "" {
				builder.WriteString(s[i])
				builder.WriteString("\n")
			}
		}
	} else {
		ipCount := len(model.Ipv4s)
		var (
			s = make([]string, ipCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range model.Ipv4s {
			go trace(c, i)
		}
	loopIPv4:
		for range s {
			select {
			case o := <-c:
				s[o.i] = o.s
			case <-t:
				break loopIPv4
			}
		}
		// 收集结果
		for _, r := range s {
			if r != "" {
				builder.WriteString(r)
				builder.WriteString("\n")
			}
		}
	}
	// 返回完整结果，去掉末尾的换行符
	result := builder.String()
	return strings.TrimSuffix(result, "\n")
}
