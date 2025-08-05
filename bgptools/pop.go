package bgptools

import (
	"fmt"
	"html"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/backtrace/model"
	"github.com/oneclickvirt/defaultset"
)

type ASCard struct {
	ASN    string
	Name   string
	Fill   string
	Stroke string
	ID     string
}

type Arrow struct {
	From string
	To   string
}

type Upstream struct {
	ASN    string
	Name   string
	Direct bool
	Tier1  bool
	Type   string
}

type PoPResult struct {
	TargetASN string
	Upstreams []Upstream
	Result    string
}

// retryConfig 重试配置
type retryConfig struct {
	maxRetries int
	timeouts   []time.Duration
}

// 默认重试配置：3次重试，超时时间分别为3s、4s、5s
var defaultRetryConfig = retryConfig{
	maxRetries: 3,
	timeouts:   []time.Duration{3 * time.Second, 4 * time.Second, 5 * time.Second},
}

// executeWithRetry 执行带重试的HTTP请求
func executeWithRetry(client *req.Client, url string, config retryConfig) (*req.Response, error) {
	var lastErr error
	for attempt := 0; attempt < config.maxRetries; attempt++ {
		timeout := config.timeouts[attempt]
		resp, err := client.SetTimeout(timeout).R().
			Get(url)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}
		if err != nil {
			lastErr = fmt.Errorf("attempt %d failed with timeout %v: %w", attempt+1, timeout, err)
		} else {
			lastErr = fmt.Errorf("attempt %d failed with HTTP status %d (timeout %v)", attempt+1, resp.StatusCode, timeout)
		}
		if attempt < config.maxRetries-1 {
			time.Sleep(3 * time.Second)
		}
	}
	return nil, fmt.Errorf("all %d attempts failed, last error: %w", config.maxRetries, lastErr)
}

func getISPAbbr(asn, name string) string {
	if abbr, ok := model.Tier1Global[asn]; ok {
		return abbr
	}
	if idx := strings.Index(name, " "); idx != -1 && idx >= 18 {
		return name[:idx]
	}
	return strings.TrimSpace(name)
}

func getISPType(asn string, tier1 bool, direct bool) string {
	switch {
	case tier1 && model.Tier1Global[asn] != "":
		return "Tier1 Global"
	case model.Tier1Regional[asn] != "":
		return "Tier1 Regional"
	case model.Tier2[asn] != "":
		return "Tier2"
	case model.ContentProviders[asn] != "":
		return "CDN Provider"
	case model.IXPS[asn] != "":
		return "IXP"
	case direct:
		return "Direct"
	default:
		return "Indirect"
	}
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func getSVGPath(ip string) (string, error) {
	if !isValidIP(ip) {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	var lastErr error
	for attempt := 0; attempt < defaultRetryConfig.maxRetries; attempt++ {
		client := req.C().ImpersonateChrome()
		url := fmt.Sprintf("https://bgp.tools/prefix/%s#connectivity", ip)
		resp, err := executeWithRetry(client, url, defaultRetryConfig)
		if err == nil {
			body := resp.String()
			re := regexp.MustCompile(`<img[^>]+id="pathimg"[^>]+src="([^"]+)"`)
			matches := re.FindStringSubmatch(body)
			if len(matches) >= 2 {
				return matches[1], nil
			}
			lastErr = fmt.Errorf("SVG path not found for IP %s", ip)
		} else {
			lastErr = fmt.Errorf("failed to fetch BGP info for IP %s: %w", ip, err)
		}
		if attempt < defaultRetryConfig.maxRetries-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return "", fmt.Errorf("failed to get SVG path after %d retries: %w", defaultRetryConfig.maxRetries, lastErr)
}

func downloadSVG(svgPath string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < defaultRetryConfig.maxRetries; attempt++ {
		client := req.C().ImpersonateChrome()
		uuid := uuid.NewString()
		url := fmt.Sprintf("https://bgp.tools%s?%s&loggedin", svgPath, uuid)
		resp, err := executeWithRetry(client, url, defaultRetryConfig)
		if err == nil {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err == nil {
				return string(bodyBytes), nil
			}
			lastErr = fmt.Errorf("failed to read SVG response body: %w", err)
		} else {
			lastErr = fmt.Errorf("failed to download SVG: %w", err)
		}
		if attempt < defaultRetryConfig.maxRetries-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return "", fmt.Errorf("failed to download SVG after %d retries: %w", defaultRetryConfig.maxRetries, lastErr)
}

func parseASAndEdges(svg string) ([]ASCard, []Arrow) {
	svg = html.UnescapeString(svg)
	var nodes []ASCard
	var edges []Arrow
	nodeRE := regexp.MustCompile(`(?s)<g id="node\d+" class="node">(.*?)</g>`)
	edgeRE := regexp.MustCompile(`(?s)<g id="edge\d+" class="edge">(.*?)</g>`)
	asnRE := regexp.MustCompile(`<title>AS(\d+)</title>`)
	nameRE := regexp.MustCompile(`xlink:title="([^"]+)"`)
	fillRE := regexp.MustCompile(`<polygon[^>]+fill="([^"]+)"`)
	strokeRE := regexp.MustCompile(`<polygon[^>]+stroke="([^"]+)"`)
	titleRE := regexp.MustCompile(`<title>AS(\d+)->AS(\d+)</title>`)
	for _, match := range nodeRE.FindAllStringSubmatch(svg, -1) {
		block := match[1]
		asn := ""
		if a := asnRE.FindStringSubmatch(block); len(a) > 1 {
			asn = a[1]
		}
		name := "unknown"
		if n := nameRE.FindStringSubmatch(block); len(n) > 1 {
			name = strings.TrimSpace(n[1])
		}
		fill := "none"
		if f := fillRE.FindStringSubmatch(block); len(f) > 1 {
			fill = f[1]
		}
		stroke := "none"
		if s := strokeRE.FindStringSubmatch(block); len(s) > 1 {
			stroke = s[1]
		}
		if asn != "" {
			nodes = append(nodes, ASCard{
				ASN:    asn,
				Name:   name,
				Fill:   fill,
				Stroke: stroke,
				ID:     "",
			})
		}
	}
	for _, match := range edgeRE.FindAllStringSubmatch(svg, -1) {
		block := match[1]
		if t := titleRE.FindStringSubmatch(block); len(t) == 3 {
			edges = append(edges, Arrow{
				From: t[1],
				To:   t[2],
			})
		}
	}
	return nodes, edges
}

func findTargetASN(nodes []ASCard) string {
	for _, n := range nodes {
		if n.Fill == "limegreen" || n.Stroke == "limegreen" || n.Fill == "green" {
			return n.ASN
		}
	}
	if len(nodes) > 0 {
		return nodes[0].ASN
	}
	return ""
}

func findUpstreams(targetASN string, nodes []ASCard, edges []Arrow) []Upstream {
	upstreamMap := map[string]bool{}
	for _, e := range edges {
		if e.From == targetASN {
			upstreamMap[e.To] = true
		}
	}
	var upstreams []Upstream
	addedASNs := map[string]bool{}
	for _, n := range nodes {
		if !upstreamMap[n.ASN] {
			continue
		}
		isTier1 := (n.Fill == "white" && n.Stroke == "#005ea5")
		upstreamType := getISPType(n.ASN, isTier1, true)
		upstreams = append(upstreams, Upstream{
			ASN:    n.ASN,
			Name:   n.Name,
			Direct: true,
			Tier1:  isTier1,
			Type:   upstreamType,
		})
		addedASNs[n.ASN] = true
	}
	if len(upstreams) == 1 {
		currentASN := upstreams[0].ASN
		for {
			nextUpstreams := map[string]bool{}
			for _, e := range edges {
				if e.From == currentASN {
					nextUpstreams[e.To] = true
				}
			}
			if len(nextUpstreams) != 1 {
				break
			}
			var nextASN string
			for asn := range nextUpstreams {
				nextASN = asn
				break
			}
			if addedASNs[nextASN] {
				break
			}
			var nextNode *ASCard
			for _, n := range nodes {
				if n.ASN == nextASN {
					nextNode = &n
					break
				}
			}
			if nextNode == nil {
				break
			}
			isTier1 := (nextNode.Fill == "white" && nextNode.Stroke == "#005ea5")
			upstreamType := getISPType(nextNode.ASN, isTier1, false)
			upstreams = append(upstreams, Upstream{
				ASN:    nextNode.ASN,
				Name:   nextNode.Name,
				Direct: false,
				Tier1:  isTier1,
				Type:   upstreamType,
			})
			addedASNs[nextNode.ASN] = true
			currentASN = nextASN
		}
	} else if len(upstreams) > 1 {
		for _, directUpstream := range upstreams {
			currentASN := directUpstream.ASN
			for {
				nextUpstreams := map[string]bool{}
				for _, e := range edges {
					if e.From == currentASN {
						nextUpstreams[e.To] = true
					}
				}
				if len(nextUpstreams) != 1 {
					break
				}
				var nextASN string
				for asn := range nextUpstreams {
					nextASN = asn
					break
				}
				if addedASNs[nextASN] {
					break
				}
				var nextNode *ASCard
				for _, n := range nodes {
					if n.ASN == nextASN {
						nextNode = &n
						break
					}
				}
				if nextNode == nil {
					break
				}
				isTier1 := (nextNode.Fill == "white" && nextNode.Stroke == "#005ea5")
				if isTier1 {
					upstreamType := getISPType(nextNode.ASN, isTier1, false)
					upstreams = append(upstreams, Upstream{
						ASN:    nextNode.ASN,
						Name:   nextNode.Name,
						Direct: false,
						Tier1:  isTier1,
						Type:   upstreamType,
					})
					addedASNs[nextNode.ASN] = true
					break
				}
				currentASN = nextASN
			}
		}
	}
	return upstreams
}

func GetPoPInfo(ip string) (*PoPResult, error) {
	if ip == "" {
		return nil, fmt.Errorf("IP address cannot be empty")
	}

	svgPath, err := getSVGPath(ip)
	if err != nil {
		return nil, fmt.Errorf("获取SVG路径失败: %w", err)
	}
	svg, err := downloadSVG(svgPath)
	if err != nil {
		return nil, fmt.Errorf("下载SVG失败: %w", err)
	}
	nodes, edges := parseASAndEdges(svg)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到任何AS节点")
	}
	targetASN := findTargetASN(nodes)
	if targetASN == "" {
		return nil, fmt.Errorf("无法识别目标 ASN")
	}
	upstreams := findUpstreams(targetASN, nodes, edges)
	colWidth := 18
	center := func(s string) string {
		runeLen := len([]rune(s))
		if runeLen >= colWidth {
			return string([]rune(s)[:colWidth])
		}
		padding := colWidth - runeLen
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}
	var result strings.Builder
	perLine := 5
	for i := 0; i < len(upstreams); i += perLine {
		end := i + perLine
		if end > len(upstreams) {
			end = len(upstreams)
		}
		batch := upstreams[i:end]
		var line1, line2, line3 []string
		for _, u := range batch {
			abbr := getISPAbbr(u.ASN, u.Name)
			asStr := center("AS" + u.ASN)
			abbrStr := center(abbr)
			typeStr := center(u.Type)
			line1 = append(line1, defaultset.White(asStr))
			line2 = append(line2, abbrStr)
			line3 = append(line3, defaultset.Blue(typeStr))
		}
		result.WriteString(strings.Join(line1, ""))
		result.WriteString("\n")
		result.WriteString(strings.Join(line2, ""))
		result.WriteString("\n")
		result.WriteString(strings.Join(line3, ""))
		result.WriteString("\n")
	}
	return &PoPResult{
		TargetASN: targetASN,
		Upstreams: upstreams,
		Result:    result.String(),
	}, nil
}
