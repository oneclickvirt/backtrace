package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
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

type ConcurrentResults struct {
	bgpResult       string
	backtraceResult string
	bgpError        error
	// backtraceError  error
}

func main() {
	go func() {
		http.Get("https://hits.spiritlhl.net/backtrace.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	fmt.Println(Green("Repo:"), Yellow("https://github.com/oneclickvirt/backtrace"))
	var showVersion, showIpInfo, help, ipv6 bool
	var specifiedIP string
	backtraceFlag := flag.NewFlagSet("backtrace", flag.ContinueOnError)
	backtraceFlag.BoolVar(&help, "h", false, "Show help information")
	backtraceFlag.BoolVar(&showVersion, "v", false, "Show version")
	backtraceFlag.BoolVar(&showIpInfo, "s", true, "Disabe show ip info")
	backtraceFlag.BoolVar(&model.EnableLoger, "log", false, "Enable logging")
	backtraceFlag.BoolVar(&ipv6, "ipv6", false, "Enable ipv6 testing")
	backtraceFlag.StringVar(&specifiedIP, "ip", "", "Specify IP address for bgptools")
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
			fmt.Printf("get ip info err %v \n", err.Error())
		} else {
			err = json.NewDecoder(rsp.Body).Decode(&info)
			if err != nil {
				fmt.Printf("json decode err %v \n", err.Error())
			} else {
				fmt.Println(Green("国家: ") + White(info.Country) + Green(" 城市: ") + White(info.City) +
					Green(" 服务商: ") + Blue(info.Org))
			}
		}
	}
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	if !preCheck.Connected {
		fmt.Println(Red("PreCheck IP Type Failed"))
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
		return
	}
	var useIPv6 bool
	switch preCheck.StackType {
	case "DualStack":
		useIPv6 = ipv6
	case "IPv4":
		useIPv6 = false
	case "IPv6":
		useIPv6 = true
	default:
		fmt.Println(Red("PreCheck IP Type Failed"))
		if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
		return
	}
	results := ConcurrentResults{}
	var wg sync.WaitGroup
	var targetIP string
	if specifiedIP != "" {
		targetIP = specifiedIP
	} else if info.Ip != "" {
		targetIP = info.Ip
	}
	if targetIP != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2; i++ {
				result, err := bgptools.GetPoPInfo(targetIP)
				results.bgpError = err
				if err == nil && result.Result != "" {
					results.bgpResult = result.Result
					return
				}
				if i == 0 {
					time.Sleep(3 * time.Second)
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := backtrace.BackTrace(useIPv6)
		results.backtraceResult = result
	}()
	wg.Wait()
	if results.bgpResult != "" {
		fmt.Print(results.bgpResult)
	}
	if results.backtraceResult != "" {
		fmt.Printf("%s\n", results.backtraceResult)
	}
	fmt.Println(Yellow("准确线路自行查看详细路由，本测试结果仅作参考"))
	fmt.Println(Yellow("同一目标地址多个线路时，检测可能已越过汇聚层，除第一个线路外，后续信息可能无效"))
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
