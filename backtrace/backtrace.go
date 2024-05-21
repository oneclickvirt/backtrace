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
	)
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
		go trace(c, i)
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
