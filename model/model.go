package model

import "time"

const BackTraceVersion = "v0.0.7"

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

var Tier1Global = map[string]string{
	"174":   "Cogent",
	"1299":  "Arelion",
	"3356":  "Lumen",
	"3257":  "GTT",
	"7018":  "AT&T",
	"701":   "Verizon",
	"2914":  "NTT",
	"6453":  "Tata",
	"3320":  "DTAG",
	"5511":  "Orange",
	"3491":  "PCCW",
	"6461":  "Zayo",
	"6830":  "Liberty",
	"6762":  "Sparkle",
	"12956": "Telxius",
	"702":   "Verizon",
}

var Tier1Regional = map[string]string{
	"4134":  "ChinaNet",
	"4837":  "China Unicom",
	"9808":  "China Mobile",
	"4766":  "Korea Telecom",
	"2516":  "KDDI",
	"7713":  "Telkomnet",
	"9121":  "Etisalat",
	"7473":  "SingTel",
	"4637":  "Telstra",
	"5400":  "British Telecom",
	"2497":  "IIJ",
	"3462":  "Chunghwa Telecom",
	"3463":  "TWNIC",
	"12389": "SoftBank",
	"3303":  "MTS",
	"45609": "Reliance Jio",
}

var Tier2 = map[string]string{
	"6939":  "HurricaneElectric",
	"20485": "Transtelecom",
	"1273":  "Vodafone",
	"1239":  "Sprint",
	"6453":  "Tata",
	"6762":  "Sparkle",
	"9002":  "RETN",
	"7922":  "Comcast",
	"23754": "Rostelecom",
	"3320":  "DTAG",
}

var ContentProviders = map[string]string{
	"15169": "Google",
	"32934": "Facebook",
	"54113": "Fastly",
	"20940": "Akamai",
	"13335": "Cloudflare",
	"14618": "Amazon AWS",
	"55102": "Netflix CDN",
	"4685":  "CacheFly",
	"16509": "Amazon",
	"36040": "Amazon CloudFront",
	"36459": "EdgeCast",
	"24940": "CDNetworks",
}

var IXPS = map[string]string{
	"5539":   "IX.br",
	"25291":  "HKIX",
	"1200":   "AMS-IX",
	"6695":   "DE-CIX",
	"58558":  "LINX",
	"395848": "France-IX",
	"4713":   "JPNAP",
	"4635":   "SIX",
	"2906":   "MSK-IX",
	"1273":   "NIX.CZ",
}
