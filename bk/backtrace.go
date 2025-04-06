package backtrace

import (
	"fmt"
	"time"
)

func BackTrace(test bool) {
	if !test {
		// 获取IP地址数量
		ipCount := len(ipv4s)
		var (
			s = make([]string, ipCount) // 动态分配切片大小
			c = make(chan Result)
			t = time.After(time.Second * 10)
		)
		for i := range ipv4s {
			go trace(c, i)
		}
	loop:
		for range s {
			select {
			case o := <-c:
				s[o.i] = o.s
			case <-t:
				break loop
			}
		}
		for _, r := range s {
			if r != "" {
				fmt.Println(r)
			}
		}
	}
}
