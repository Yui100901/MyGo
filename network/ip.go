package main

import (
	"fmt"
	"net"
)

//
// @Author yfy2001
// @Date 2025/2/28 09 10
//

// NetInfo 定义 NetInfo 结构体
type NetInfo struct {
	Name          string   // 接口名称
	HardwareAddr  string   // 硬件地址
	IPv4Addresses []string // IPv4 地址列表（CIDR 格式）
	IPv6Addresses []string // IPv6 地址列表（CIDR 格式）
}

func NewNetInfo() ([]*NetInfo, error) {
	var netInfos []*NetInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	for _, iface := range interfaces {
		// 跳过未启用的和环回接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		netInfo := &NetInfo{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
		}

		adders, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Error getting addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		// 遍历地址
		for _, addr := range adders {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				if ip.To4() != nil {
					// 添加到 IPv4 列表
					netInfo.IPv4Addresses = append(netInfo.IPv4Addresses, v.String())
				} else {
					// 添加到 IPv6 列表
					netInfo.IPv6Addresses = append(netInfo.IPv6Addresses, v.String())
				}
			case *net.IPAddr:
				ip := v.IP
				if ip.To4() != nil {
					// 添加到 IPv4 列表
					netInfo.IPv4Addresses = append(netInfo.IPv4Addresses, v.String())
				} else {
					// 添加到 IPv6 列表
					netInfo.IPv6Addresses = append(netInfo.IPv6Addresses, v.String())
				}
			}
		}

		netInfos = append(netInfos, netInfo)
	}
	return netInfos, nil
}
