package backtrace

import (
	"net"
)

// GeneratePrefixList 生成指定前缀的IP地址列表
func GeneratePrefixList(prefix string) []string {
	// 解析CIDR表示法的IP地址
	_, ipNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil
	}
	// 获取IP地址的32位整数表示
	ip := ipNet.IP.To4()
	start := binaryIPToInt(ip)
	maskSize, _ := ipNet.Mask.Size()
	end := start | (1<<(32-maskSize) - 1)
	// 生成IP地址列表
	var prefixList []string
	for i := start; i <= end; i++ {
		if (i-start)%256 == 0 {
			tempText := intToBinaryIP(i).String()
			prefixList = append(prefixList, tempText[:len(tempText)-2])
		}
	}
	return prefixList
}

// 将IP地址转换为32位整数
func binaryIPToInt(ip net.IP) uint32 {
	return (uint32(ip[0]) << 24) | (uint32(ip[1]) << 16) | (uint32(ip[2]) << 8) | uint32(ip[3])
}

// 将32位整数转换为IP地址
func intToBinaryIP(i uint32) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}
