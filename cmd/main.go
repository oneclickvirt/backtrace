package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
	"github.com/oneclickvirt/backtrace/model"
	"github.com/oneclickvirt/backtrace/utils"
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
		http.Get("https://hits.spiritlhl.net/backtrace.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	fmt.Println(Green("Repo:"), Yellow("https://github.com/oneclickvirt/backtrace"))
	var showVersion, showIpInfo, help, ipv6 bool
	backtraceFlag := flag.NewFlagSet("backtrace", flag.ContinueOnError)
	backtraceFlag.BoolVar(&help, "h", false, "Show help information")
	backtraceFlag.BoolVar(&showVersion, "v", false, "Show version")
	backtraceFlag.BoolVar(&showIpInfo, "s", true, "Disabe show ip info")
	backtraceFlag.BoolVar(&model.EnableLoger, "log", false, "Enable logging")
	backtraceFlag.BoolVar(&ipv6, "ipv6", false, "Enable ipv6 testing")
	backtraceFlag.Parse(os.Args[1:])
	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		backtraceFlag.PrintDefaults()
		return
	}
	if showVersion {
		fmt.Println(model.BackTraceVersion)
		return
	}
	info := IpInfo{}
	if showIpInfo {
		rsp, err := http.Get("http://ipinfo.io")
		if err != nil {
			fmt.Errorf("Get ip info err %v \n", err.Error())
		} else {
			err = json.NewDecoder(rsp.Body).Decode(&info)
			if err != nil {
				fmt.Errorf("json decode err %v \n", err.Error())
			} else {
				fmt.Println(Green("国家: ") + White(info.Country) + Green(" 城市: ") + White(info.City) +
					Green(" 服务商: ") + Blue(info.Org))
			}
		}
	}
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	if preCheck.Connected && info.Ip != "" {
		result, err := bgptools.GetPoPInfo(info.Ip)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Print(result.Result)
	}
	if preCheck.Connected && preCheck.StackType == "DualStack" {
		backtrace.BackTrace(ipv6)
	} else if preCheck.Connected && preCheck.StackType == "IPv4" {
		backtrace.BackTrace(false)
	} else if preCheck.Connected && preCheck.StackType == "IPv6" {
		backtrace.BackTrace(true)
	} else {
		fmt.Println(Red("PreCheck IP Type Failed"))
	}
	fmt.Println(Yellow("准确线路自行查看详细路由，本测试结果仅作参考"))
	fmt.Println(Yellow("同一目标地址多个线路时，可能检测已越过汇聚层，除了第一个线路外，后续信息可能无效"))
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
