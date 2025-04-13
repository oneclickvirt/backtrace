package model

import "time"

const BackTraceVersion = "v0.0.5"

var EnableLoger = false

// IcmpTarget 定义ICMP目标的JSON结构
type IcmpTarget struct {
	Province  string `json:"province"`
	ISP       string `json:"isp"`
	IPVersion string `json:"ip_version"`
	IPs       string `json:"ips"` // IP列表，以逗号分隔
}

var (
	IcmpTargets = "https://raw.githubusercontent.com/spiritLHLS/icmp_targets/main/nodes.json"
	CdnList     = []string{
		"http://cdn1.spiritlhl.net/",
		"http://cdn2.spiritlhl.net/",
		"http://cdn3.spiritlhl.net/",
		"http://cdn4.spiritlhl.net/",
	}
	Ipv4s = []string{
		"219.141.140.10",  // 北京电信v4
		"202.106.195.68",  // 北京联通v4
		"221.179.155.161", // 北京移动v4
		"202.96.209.133",  // 上海电信v4
		"210.22.97.1",     // 上海联通v4
		"211.136.112.200", // 上海移动v4
		"58.60.188.222",   // 广州电信v4
		"210.21.196.6",    // 广州联通v4
		"120.196.165.24",  // 广州移动v4
		"61.139.2.69",     // 成都电信v4
		"119.6.6.6",       // 成都联通v4
		"211.137.96.205",  // 成都移动v4
	}

	Ipv4Names = []string{
		"北京电信v4", "北京联通v4", "北京移动v4",
		"上海电信v4", "上海联通v4", "上海移动v4",
		"广州电信v4", "广州联通v4", "广州移动v4",
		"成都电信v4", "成都联通v4", "成都移动v4",
	}
	Ipv6s = []string{
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
	Ipv6Names = []string{
		"北京电信v6", "北京联通v6", "北京移动v6",
		"上海电信v6", "上海联通v6", "上海移动v6",
		"广州电信v6", "广州联通v6", "广州移动v6",
	}
	M = map[string]string{
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
	CachedIcmpData          string
	CachedIcmpDataFetchTime time.Time
	ParsedIcmpTargets       []IcmpTarget
)
