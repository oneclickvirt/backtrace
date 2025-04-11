package backtrace

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/backtrace/model"

	. "github.com/oneclickvirt/defaultset"
)

type Result struct {
	i int
	s string
}

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

// getData 获取目标地址的文本内容
func getData(endpoint string) string {
	client := req.C()
	client.SetTimeout(10 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	for _, baseUrl := range model.CdnList {
		url := baseUrl + endpoint
		resp, err := client.R().Get(url)
		if err == nil {
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if strings.Contains(string(b), "error") {
				continue
			}
			if err == nil {
				if model.EnableLoger {
					Logger.Info(fmt.Sprintf("Received data length: %d", len(b)))
				}
				return string(b)
			}
		}
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("HTTP request failed: %v", err))
		}
	}
	return ""
}

// tryAlternativeIPs 从IcmpTargets获取备选IP地址
func tryAlternativeIPs(targetName string, ipVersion string) []string {
	jsonData := getData(model.IcmpTargets)
	if jsonData == "" {
		return nil
	}
	// 简单解析JSON，提取省份和ISP信息
	var targetProvince, targetISP string
	// 从目标名称中提取省份和ISP信息
	if strings.Contains(targetName, "北京") {
		targetProvince = "北京"
	} else if strings.Contains(targetName, "上海") {
		targetProvince = "上海"
	} else if strings.Contains(targetName, "广州") {
		targetProvince = "广东"
	} else if strings.Contains(targetName, "成都") {
		targetProvince = "四川"
	}
	if strings.Contains(targetName, "电信") {
		targetISP = "电信"
	} else if strings.Contains(targetName, "联通") {
		targetISP = "联通"
	} else if strings.Contains(targetName, "移动") {
		targetISP = "移动"
	}
	// 如果没有提取到信息，返回空
	if targetProvince == "" || targetISP == "" {
		return nil
	}
	// 解析JSON数据寻找匹配的记录
	var result []string
	for _, line := range strings.Split(jsonData, "},{") {
		if strings.Contains(line, "\"province\":\""+targetProvince+"省\"") &&
			strings.Contains(line, "\"isp\":\""+targetISP+"\"") &&
			strings.Contains(line, "\"ip_version\":\""+ipVersion+"\"") {
			// 提取IP列表
			ipsStart := strings.Index(line, "\"ips\":\"") + 7
			if ipsStart > 7 {
				ipsEnd := strings.Index(line[ipsStart:], "\"")
				if ipsEnd > 0 {
					ipsList := line[ipsStart : ipsStart+ipsEnd]
					ips := strings.Split(ipsList, ",")
					// 最多返回3个不重复的IP地址
					count := 0
					for _, ip := range ips {
						if ip != "" {
							result = append(result, ip)
							count++
							if count >= 3 {
								break
							}
						}
					}
					return result
				}
			}
		}
	}
	return nil
}
