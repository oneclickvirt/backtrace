package bgptools

import (
	"fmt"
	"testing"
)

func TestGetPoPInfo(t *testing.T) {
	result, err := GetPoPInfo("66.70.153.71")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("目标 ASN: %s\n", result.TargetASN)
	fmt.Println("上游信息:")
	fmt.Print(result.Result)
}
