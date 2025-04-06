package backtrace

import (
	"fmt"
	"time"
)

func BackTrace(test bool) {
	if test {
		ipv4Count := len(ipv4s)
		ipv6Count := len(ipv6s)
		totalCount := ipv4Count + ipv6Count
		var (
			s = make([]string, totalCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range ipv4s {
			go trace(c, i)
		}
		for i := range ipv6s {
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
		ipCount := len(ipv4s)
		var (
			s = make([]string, ipCount)
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range ipv4s {
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
