package backtrace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/oneclickvirt/backtrace/defaultset"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func BackTrace() {
	var (
		s     [12]string // 对应 ips 目标地址数量
		c     = make(chan Result)
		t     = time.After(time.Second * 10)
		cmin2 = []string{
			// 以下均为 /24 地址
			"223.118.32",
			"223.119.32", "223.119.34", "223.119.35", "223.119.36", "223.119.37", "223.119.100", "223.119.253",
			"223.120.165"}
		// 其他区间的地址
		prefixes = []string{
			"223.119.8.0/21",
			"223.120.128.0/17",
			"223.120.134/23",
			"223.120.136/23",
			"223.120.138/23",
			"223.120.154/23",
			"223.120.158/23",
			"223.120.164/22",
			"223.120.168/22",
			"223.120.172/22",
			"223.120.174/23",
			"223.120.184/22",
			"223.120.188/22",
			"223.120.192/23",
			"223.120.200/23",
			"223.120.210/23",
			"223.120.212/23",
		}
	)
	// 生成CMIN2的IPV4前缀地址
	for _, prefix := range prefixes {
		cmin2 = append(cmin2, GeneratePrefixList(prefix)...)
	}
	rsp, err := http.Get("http://ipinfo.io")
	if err != nil {
		fmt.Errorf("Get ip info err %v \n", err.Error())
	} else {
		info := IpInfo{}
		err = json.NewDecoder(rsp.Body).Decode(&info)
		if err != nil {
			fmt.Errorf("json decode err %v \n", err.Error())
		} else {
			fmt.Println(Green("国家: ") + White(info.Country) + Green(" 城市: ") + White(info.City) +
				Green(" 服务商: ") + Blue(info.Org))
		}
	}
	for i := range ips {
		go trace(c, i, cmin2)
	}
loop:
	for range s {
		select {
		case o := <-c:
			s[o.i] = o.s
		case <-t:
			break loop
		}
	}
	for _, r := range s {
		fmt.Println(r)
	}
}
