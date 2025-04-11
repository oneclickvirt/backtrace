package backtrace

import (
	"encoding/json"
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
	client.SetTimeout(6 * time.Second)
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

// parseIcmpTargets 解析ICMP目标数据
func parseIcmpTargets(jsonData string) []model.IcmpTarget {
	// 确保JSON数据格式正确，如果返回的是数组，需要添加[和]
	if !strings.HasPrefix(jsonData, "[") {
		jsonData = "[" + jsonData + "]"
	}
	// 如果JSON数据中的对象没有正确用逗号分隔，修复它
	jsonData = strings.ReplaceAll(jsonData, "}{", "},{")
	var targets []model.IcmpTarget
	err := json.Unmarshal([]byte(jsonData), &targets)
	if err != nil {
		if model.EnableLoger {
			Logger.Error(fmt.Sprintf("Failed to parse ICMP targets: %v", err))
		}
		return nil
	}
	return targets
}

// tryAlternativeIPs 从IcmpTargets获取备选IP地址
func tryAlternativeIPs(targetName string, ipVersion string) []string {
	if model.ParsedIcmpTargets == nil || (model.ParsedIcmpTargets != nil && len(model.ParsedIcmpTargets) == 0) {
		return nil
	}
	// 从目标名称中提取省份和ISP信息
	var targetProvince, targetISP string
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
	// 查找匹配条件的目标
	var result []string
	for _, target := range model.ParsedIcmpTargets {
		// 检查省份是否匹配（可能带有"省"字或不带）
		provinceMatch := (target.Province == targetProvince) || (target.Province == targetProvince+"省")
		// 检查ISP和IP版本是否匹配
		if provinceMatch && target.ISP == targetISP && target.IPVersion == ipVersion {
			// 解析IP列表
			if target.IPs != "" {
				ips := strings.Split(target.IPs, ",")
				// 最多返回3个IP地址
				count := 0
				for _, ip := range ips {
					if ip != "" {
						result = append(result, strings.TrimSpace(ip))
						count++
						if count >= 3 {
							break
						}
					}
				}
				if len(result) > 0 {
					return result
				}
			}
		}
	}
	return nil
}
