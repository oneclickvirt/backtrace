package backtrace

import (
	"strings"
)

type Result struct {
	i int
	s string
}

var (
	ipv4s = []string{
		// "219.141.136.12", "202.106.50.1",
		"219.141.140.10", "202.106.195.68", "221.179.155.161",
		"202.96.209.133", "210.22.97.1", "211.136.112.200",
		"58.60.188.222", "210.21.196.6", "120.196.165.24",
		"61.139.2.69", "119.6.6.6", "211.137.96.205",
	}
	ipv4Names = []string{
		"北京电信v4", "北京联通v4", "北京移动v4",
		"上海电信v4", "上海联通v4", "上海移动v4",
		"广州电信v4", "广州联通v4", "广州移动v4",
		"成都电信v4", "成都联通v4", "成都移动v4",
	}
	ipv6s = []string{
		"2400:89c0:1053:3::69",    // 北京电信 IPv6
		"2400:89c0:1013:3::54",    // 北京联通 IPv6
		"2409:8c00:8421:1303::55", // 北京移动 IPv6
		"240e:e1:aa00:4000::24",   // 上海电信 IPV6
		"2408:80f1:21:5003::a",    // 上海联通 IPv6
		"2409:8c1e:75b0:3003::26", // 上海移动 IPv6
		"240e:97c:2f:3000::44",    // 广州电信 IPv6
		"2408:8756:f50:1001::c",   // 广州联通 IPv6
		"2409:8c54:871:1001::12",  // 广州移动 IPv6
	}
	ipv6Names = []string{
		"北京电信v6", "北京联通v6", "北京移动v6",
		"上海电信v6", "上海联通v6", "上海移动v6",
		"广州电信v6", "广州联通v6", "广州移动v6",
	}
	m = map[string]string{
		// [] 前的字符串个数，中文占2个字符串
		"AS23764": "电信CTGNET [精品线路]",
		"AS4809a": "电信CN2GIA [精品线路]",
		"AS4809b": "电信CN2GT  [优质线路]",
		"AS4809":  "电信CN2    [优质线路]",
		"AS4134":  "电信163    [普通线路]",
		"AS9929":  "联通9929   [优质线路]",
		"AS4837":  "联通4837   [普通线路]",
		"AS58807": "移动CMIN2  [精品线路]",
		"AS9808":  "移动CMI    [普通线路]",
		"AS58453": "移动CMI    [普通线路]",
	}
)

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{} // 用于存储已经遇到的元素
	result := []string{}             // 存储去重后的结果
	for v := range elements {        // 遍历切片中的元素
		if encountered[elements[v]] == true { // 如果该元素已经遇到过
			// 存在过就不加入了
		} else {
			encountered[elements[v]] = true      // 将该元素标记为已经遇到
			result = append(result, elements[v]) // 将该元素加入到结果切片中
		}
	}
	return result // 返回去重后的结果切片
}

func ipAsn(ip string) string {
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
