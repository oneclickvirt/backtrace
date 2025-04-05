package backtrace

import (
	"strings"
)

// 识别IPv6地址的ASN
func ipv6Asn(ip string) string {
	switch {
	// 电信CN2GIA
	case strings.HasPrefix(ip, "2408:80"):
		return "AS4809a"
	// 电信CN2GT
	case strings.HasPrefix(ip, "2408:8000"):
		return "AS4809b"
	// 电信163
	case strings.HasPrefix(ip, "240e:") || strings.HasPrefix(ip, "2408:8"):
		return "AS4134"
	// 联通9929
	case strings.HasPrefix(ip, "2408:8026:"):
		return "AS9929"
	// 联通4837
	case strings.HasPrefix(ip, "2408:8000:"):
		return "AS4837"
	// 移动CMIN2
	case strings.HasPrefix(ip, "2409:8880:"):
		return "AS58807"
	// 移动CMI
	case strings.HasPrefix(ip, "2409:8000:") || strings.HasPrefix(ip, "2409:"):
		return "AS9808"
	// 移动CMI
	case strings.HasPrefix(ip, "2407:") || strings.HasPrefix(ip, "2401:"):
		return "AS58453"
	// 电信CTGNET
	case strings.HasPrefix(ip, "2402:0:") || strings.HasPrefix(ip, "2400:8:"):
		return "AS23764"
	default:
		return ""
	}
}
