package bgptools

import (
	"fmt"
	"testing"
)

func TestGetPoPInfo(t *testing.T) {
	result, err := GetPoPInfo("206.190.233.1")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("目标 ASN: %s\n", result.TargetASN)
	fmt.Println(len(result.Upstreams))
	fmt.Println(result.Upstreams)
	fmt.Println("上游信息:")
	fmt.Print(result.Result)
}
