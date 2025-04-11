package backtrace

import (
	"fmt"
	"time"

	"github.com/oneclickvirt/backtrace/model"
)

func BackTrace(enableIpv6 bool) {
	if enableIpv6 {
		ipv4Count := len(model.Ipv4s)
		ipv6Count := len(model.Ipv6s)
		totalCount := ipv4Count + ipv6Count
		var (
			s = make([]string, totalCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range model.Ipv4s {
			go trace(c, i)
		}
		for i := range model.Ipv6s {
			go traceIPv6(c, i, ipv4Count)
		}
	loopIPv4v6:
		for range s {
			select {
			case o := <-c:
				s[o.i] = o.s
			case <-t:
				break loopIPv4v6
			}
		}
		for i := 0; i < ipv4Count; i++ {
			if s[i] != "" {
				fmt.Println(s[i])
			}
		}
		for i := ipv4Count; i < totalCount; i++ {
			if s[i] != "" {
				fmt.Println(s[i])
			}
		}
	} else {
		ipCount := len(model.Ipv4s)
		var (
			s = make([]string, ipCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range model.Ipv4s {
			go trace(c, i)
		}
	loopIPv4:
		for range s {
			select {
			case o := <-c:
				s[o.i] = o.s
			case <-t:
				break loopIPv4
			}
		}
		for _, r := range s {
			if r != "" {
				fmt.Println(r)
			}
		}
	}
}
