package backtrace

import (
	"fmt"
	"net"
	"strings"

	. "github.com/oneclickvirt/backtrace/defaultset"
)

type Result struct {
	i int
	s string
}

var (
	ips = []string{
		"219.141.136.12", "202.106.50.1", "221.179.155.161",
		"202.96.209.133", "210.22.97.1", "211.136.112.200",
		"58.60.188.222", "210.21.196.6", "120.196.165.24",
		"61.139.2.69", "119.6.6.6", "211.137.96.205",
	}
	names = []string{
		"北京电信", "北京联通", "北京移动",
		"上海电信", "上海联通", "上海移动",
		"广州电信", "广州联通", "广州移动",
		"成都电信", "成都联通", "成都移动",
	}
	m = map[string]string{
		// [] 前的字符串个数，中文占2个字符串
		"AS4809a": " 电信CN2GIA [优质线路]", // 18
		"AS4809b": " 电信CN2GT  [精品线路]", // 18
		"AS4134":  " 电信163    [普通线路]", // 18
		"AS9929":  " 联通9929   [优质线路]", // 18
		"AS4837":  " 联通4837   [普通线路]", // 18
		"AS58807": "移动CMIN2  [精品线路]",  // 18
		"AS9808":  " 移动CMI    [普通线路]", // 18
		"AS58453": "移动CMI    [普通线路]",  // 18
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

func trace(ch chan Result, i int, cmin2 []string) {
	hops, err := Trace(net.ParseIP(ips[i]))
	if err != nil {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], err)
		ch <- Result{i, s}
		return
	}
	var asns []string
	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipAsn(n.IP.String(), cmin2)
			if asn != "" {
				asns = append(asns, asn)
			}
		}
	}
	// 处理CN2不同路线的区别
	if asns != nil && len(asns) > 0 {
		var tempText string
		asns = removeDuplicates(asns)
		tempText += fmt.Sprintf("%v ", names[i])
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
		if hasAS4134 && hasAS4809 {
			// 同时包含 AS4134 和 AS4809 属于 CN2GT
			asns = append(asns, "AS4809b")
		} else if hasAS4809 {
			// 仅包含 AS4809 属于 CN2GIA
			asns = append(asns, "AS4809a")
		}
		tempText += fmt.Sprintf("%-15s ", ips[i])
		for _, asn := range asns {
			asnDescription := m[asn]
			switch asn {
			case "":
				continue
			case "AS4809": // 被 AS4809a 和 AS4809b 替代了
				continue
			case "AS9929":
				tempText += DarkGreen(asnDescription) + " "
			case "AS4809a":
				tempText += DarkGreen(asnDescription) + " "
			case "AS4809b":
				tempText += Green(asnDescription) + " "
			case "AS58807":
				tempText += Green(asnDescription) + " "
			default:
				tempText += White(asnDescription) + " "
			}
		}
		ch <- Result{i, tempText}
	} else {
		s := fmt.Sprintf("%v %-15s %v", names[i], ips[i], Red("检测不到ASN"))
		ch <- Result{i, s}
	}
}

func ipAsn(ip string, cmin2 []string) string {
	switch {
	case strings.HasPrefix(ip, "59.43"):
		return "AS4809"
	case strings.HasPrefix(ip, "202.97"):
		return "AS4134"
	case strings.HasPrefix(ip, "218.105") || strings.HasPrefix(ip, "210.51"):
		return "AS9929"
	case strings.HasPrefix(ip, "219.158"):
		return "AS4837"
	case strings.HasPrefix(ip, "223.118") || strings.HasPrefix(ip, "223.119") ||
		strings.HasPrefix(ip, "223.120"):
		for _, prefix := range cmin2 {
			if strings.HasPrefix(ip, prefix) {
				return "AS58807"
			}
		}
		return "AS58453"
	case strings.HasPrefix(ip, "103.11.109") || strings.HasPrefix(ip, "45.204.69") ||
		strings.HasPrefix(ip, "223.121"):
		return "AS58453"
	default:
		return ""
	}
}
