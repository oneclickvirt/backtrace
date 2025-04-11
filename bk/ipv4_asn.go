package backtrace

import (
	"strings"
)

func ipv4Asn(ip string) string {
	if strings.Contains(ip, ":") {
		return ipv6Asn(ip)
	}
	switch {
	case strings.HasPrefix(ip, "59.43"):
		return "AS4809"
	case strings.HasPrefix(ip, "202.97"):
		return "AS4134"
	case strings.HasPrefix(ip, "218.105") || strings.HasPrefix(ip, "210.51"):
		return "AS9929"
	case strings.HasPrefix(ip, "219.158"):
		return "AS4837"
	case strings.HasPrefix(ip, "223.120.19") || strings.HasPrefix(ip, "223.120.17") || strings.HasPrefix(ip, "223.120.16") ||
		strings.HasPrefix(ip, "223.120.140") || strings.HasPrefix(ip, "223.120.130") || strings.HasPrefix(ip, "223.120.131") ||
		strings.HasPrefix(ip, "223.120.141"):
		return "AS58807"
	case strings.HasPrefix(ip, "223.118") || strings.HasPrefix(ip, "223.119") || strings.HasPrefix(ip, "223.120") || strings.HasPrefix(ip, "223.121"):
		return "AS58453"
	case strings.HasPrefix(ip, "69.194") || strings.HasPrefix(ip, "203.22"):
		return "AS23764"
	default:
		return ""
	}
}
