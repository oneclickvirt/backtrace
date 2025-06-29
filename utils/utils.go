package utils

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

type NetCheckResult struct {
	HasIPv4   bool
	HasIPv6   bool
	Connected bool
	StackType string // "IPv4", "IPv6", "DualStack", "None"
}

func makeResolver(proto, dnsAddr string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, proto, dnsAddr)
		},
	}
}

func CheckPublicAccess(timeout time.Duration) NetCheckResult {
	if timeout < 2*time.Second {
		timeout = 2 * time.Second
	}
	var wg sync.WaitGroup
	resultChan := make(chan string, 8)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	checks := []struct {
		Tag  string
		Addr string
		Kind string // udp4, udp6, http4, http6
	}{
		// UDP DNS
		{"IPv4", "223.5.5.5:53", "udp4"},              // 阿里 DNS
		{"IPv4", "8.8.8.8:53", "udp4"},                // Google DNS
		{"IPv6", "[2400:3200::1]:53", "udp6"},         // 阿里 IPv6 DNS
		{"IPv6", "[2001:4860:4860::8888]:53", "udp6"}, // Google IPv6 DNS
		// HTTP HEAD
		{"IPv4", "https://www.baidu.com", "http4"},     // 百度
		{"IPv4", "https://1.1.1.1", "http4"},           // Cloudflare
		{"IPv6", "https://[2400:3200::1]", "http6"},    // 阿里 IPv6
		{"IPv6", "https://[2606:4700::1111]", "http6"}, // Cloudflare IPv6
	}
	for _, check := range checks {
		wg.Add(1)
		go func(tag, addr, kind string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			switch kind {
			case "udp4", "udp6":
				dialer := &net.Dialer{
					Timeout: timeout / 4,
				}
				conn, err := dialer.DialContext(ctx, kind, addr)
				if err == nil && conn != nil {
					conn.Close()
					select {
					case resultChan <- tag:
					case <-ctx.Done():
						return
					}
				}
			case "http4", "http6":
				var resolver *net.Resolver
				if kind == "http4" {
					resolver = makeResolver("udp4", "223.5.5.5:53")
				} else {
					resolver = makeResolver("udp6", "[2400:3200::1]:53")
				}
				dialer := &net.Dialer{
					Timeout:  timeout / 4,
					Resolver: resolver,
				}
				transport := &http.Transport{
					DialContext:           dialer.DialContext,
					MaxIdleConns:          1,
					MaxIdleConnsPerHost:   1,
					IdleConnTimeout:       time.Second,
					TLSHandshakeTimeout:   timeout / 4,
					ResponseHeaderTimeout: timeout / 4,
					DisableKeepAlives:     true,
				}
				client := &http.Client{
					Timeout:   timeout / 4,
					Transport: transport,
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				req, err := http.NewRequestWithContext(ctx, "HEAD", addr, nil)
				if err != nil {
					return
				}
				resp, err := client.Do(req)
				if err == nil && resp != nil {
					if resp.Body != nil {
						resp.Body.Close()
					}
					if resp.StatusCode < 500 {
						select {
						case resultChan <- tag:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(check.Tag, check.Addr, check.Kind)
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	hasV4 := false
	hasV6 := false
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				goto result
			}
			if res == "IPv4" {
				hasV4 = true
			}
			if res == "IPv6" {
				hasV6 = true
			}
		case <-ctx.Done():
			goto result
		}
	}
result:
	stack := "None"
	if hasV4 && hasV6 {
		stack = "DualStack"
	} else if hasV4 {
		stack = "IPv4"
	} else if hasV6 {
		stack = "IPv6"
	}
	return NetCheckResult{
		HasIPv4:   hasV4,
		HasIPv6:   hasV6,
		Connected: hasV4 || hasV6,
		StackType: stack,
	}
}