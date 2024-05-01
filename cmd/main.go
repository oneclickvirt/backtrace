package main

import (
	"fmt"
	"net/http"
	"github.com/oneclickvirt/backtrace/backtrace"
	. "github.com/oneclickvirt/backtrace/defaultset"
)

func main() {
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fbacktrace&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println(Green("项目地址:"), Yellow("https://github.com/oneclickvirt/backtrace"))
	backtrace.BackTrace()
}
