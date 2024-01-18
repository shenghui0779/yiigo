package util

import "net"

// IP2Long IP地址转整数
func IP2Long(ip string) uint32 {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return 0
	}

	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// Long2IP 整数转IP地址
func Long2IP(ip uint32) string {
	return net.IPv4(byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip)).String()
}
