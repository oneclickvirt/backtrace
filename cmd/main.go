package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"os"

	backtrace "github.com/oneclickvirt/backtrace/bk"
	. "github.com/oneclickvirt/defaultset"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func main() {
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fbacktrace&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println(Green("项目地址:"), Yellow("https://github.com/oneclickvirt/backtrace"))
	var showVersion, showIpInfo, help bool
	backtraceFlag := flag.NewFlagSet("backtrace", flag.ContinueOnError)
	backtraceFlag.BoolVar(&help, "h", false, "Show help information")
	backtraceFlag.BoolVar(&showVersion, "v", false, "Show version")
	backtraceFlag.BoolVar(&showIpInfo, "s", true, "Disabe show ip info")
	backtraceFlag.BoolVar(&backtrace.EnableLoger, "e", false, "Enable logging")
	backtraceFlag.Parse(os.Args[1:])
	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		backtraceFlag.PrintDefaults()
		return
	}
	if showVersion {
		fmt.Println(backtrace.BackTraceVersion)
		return
	}
	if showIpInfo {
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
	}
	backtrace.BackTrace()
	fmt.Println(Yellow("准确线路自行查看详细路由，本测试结果仅作参考"))
	fmt.Println(Yellow("同一目标地址多个线路时，可能检测已越过汇聚层，除了第一个线路外，后续信息可能无效"))
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
