package bgptools

import (
	"fmt"
	"html"
	"io"
	"net"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/imroc/req/v3"
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

// === 分类映射 ===

var tier1Global = map[string]string{
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
}

var tier1Regional = map[string]string{
	"4134": "ChinaNet",
	"4837": "China Unicom",
	"9808": "China Mobile",
	"4766": "Korea Telecom",
	"2516": "KDDI",
	"7713": "Telkomnet",
	"9121": "Etisalat",
}

var tier2 = map[string]string{
	"6939":  "Hurricane Electric",
	"20485": "Transtelecom",
	"1273":  "Vodafone",
	"1239":  "Sprint",
	"6762":  "Sparkle",
	"6453":  "Tata",
}

var contentProviders = map[string]string{
	"15169": "Google",
	"32934": "Facebook",
	"54113": "Fastly",
	"20940": "Akamai",
	"13335": "Cloudflare",
}

var ixps = map[string]string{
	"5539":  "IX.br",
	"25291": "HKIX",
	"1200":  "AMS-IX",
	"6695":  "DE-CIX",
}

// === 工具函数 ===

func getISPAbbr(asn, name string) string {
	if abbr, ok := tier1Global[asn]; ok {
		return abbr
	}
	if idx := strings.Index(name, " "); idx != -1 {
		return name[:idx]
	}
	return name
}

func getISPType(asn string, tier1 bool) string {
	switch {
	case tier1 && tier1Global[asn] != "":
		return "Tier1 Global"
	case tier1Regional[asn] != "":
		return "Tier1 Regional"
	case tier2[asn] != "":
		return "Tier2"
	case contentProviders[asn] != "":
		return "CDN Provider"
	case ixps[asn] != "":
		return "IXP"
	default:
		return "Direct"
	}
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func getSVGPath(client *req.Client, ip string) (string, error) {
	if !isValidIP(ip) {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	url := fmt.Sprintf("https://bgp.tools/prefix/%s#connectivity", ip)
	resp, err := client.R().Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch BGP info for IP %s: %w", ip, err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d when fetching BGP info for IP %s", resp.StatusCode, ip)
	}
	body := resp.String()
	re := regexp.MustCompile(`<img[^>]+id="pathimg"[^>]+src="([^"]+)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("SVG path not found for IP %s", ip)
	}
	return matches[1], nil
}

func downloadSVG(client *req.Client, svgPath string) (string, error) {
	uuid := uuid.NewString()
	url := fmt.Sprintf("https://bgp.tools%s?%s&loggedin", svgPath, uuid)
	resp, err := client.R().Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download SVG: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d when downloading SVG", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read SVG response body: %w", err)
	}
	return string(bodyBytes), nil
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
	for _, n := range nodes {
		if !upstreamMap[n.ASN] {
			continue
		}
		isTier1 := (n.Fill == "white" && n.Stroke == "#005ea5")
		upstreamType := getISPType(n.ASN, isTier1)
		upstreams = append(upstreams, Upstream{
			ASN:    n.ASN,
			Name:   n.Name,
			Direct: true,
			Tier1:  isTier1,
			Type:   upstreamType,
		})
	}
	return upstreams
}

func GetPoPInfo(ip string) (*PoPResult, error) {
	if ip == "" {
		return nil, fmt.Errorf("IP address cannot be empty")
	}
	client := req.C().ImpersonateChrome()
	svgPath, err := getSVGPath(client, ip)
	if err != nil {
		return nil, fmt.Errorf("获取SVG路径失败: %w", err)
	}
	svg, err := downloadSVG(client, svgPath)
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
	if len(upstreams) > 5 {
		upstreams = upstreams[:5]
	}
	colWidth := 16
	center := func(s string) string {
		runeLen := len([]rune(s))
		if runeLen >= colWidth {
			return s[:colWidth]
		}
		padding := colWidth - runeLen
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}
	var line1, line2, line3 []string
	for _, u := range upstreams {
		abbr := getISPAbbr(u.ASN, u.Name)
		line1 = append(line1, center("AS"+u.ASN))
		line2 = append(line2, center(abbr))
		line3 = append(line3, center(u.Type))
	}
	var result strings.Builder
	result.WriteString(strings.Join(line1, ""))
	result.WriteString("\n")
	result.WriteString(strings.Join(line2, ""))
	result.WriteString("\n")
	result.WriteString(strings.Join(line3, ""))
	return &PoPResult{
		TargetASN: targetASN,
		Upstreams: upstreams,
		Result:    result.String(),
	}, nil
}
