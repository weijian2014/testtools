package client

import (
	"fmt"
	"net"
	"strings"
	"testtools/common"
)

var (
	sendToServerIpAddress          string
	clientBindIpAddressRange       []string
	clientBindIpAddressRangeLength uint64 = 0
)

func StartClient() {
	err := checkJsonConfig()
	if nil != err {
		panic(err)
	}

	err = checkFlags()
	if nil != err {
		panic(err)
	}

	err = parsePort()
	if nil != err {
		panic(err)
	}

	localAddr := &common.IpAndPort{Ip: common.FlagInfos.ClientBindIpAddress, Port: 0}
	remoteAddr := &common.IpAndPort{Ip: sendToServerIpAddress, Port: common.FlagInfos.SentToServerPort}
	// Tcp
	if common.FlagInfos.UsingTcp {
		sendByTcp(localAddr, remoteAddr)
		return
	}

	// Udp
	if common.FlagInfos.UsingUdp {
		sendByUdp(localAddr, remoteAddr)
		return
	}

	// Http
	if common.FlagInfos.UsingHttp {
		sendByHttp(localAddr, remoteAddr)
		return
	}

	// Https
	if common.FlagInfos.UsingHttps {
		sendByHttps(localAddr, remoteAddr)
		return
	}

	// Quic
	if common.FlagInfos.UsingQuic {
		sendByQuic(localAddr, remoteAddr)
		return
	}

	// Dns
	if common.FlagInfos.UsingDns {
		sendByDns(localAddr, remoteAddr)
		return
	}
}

func checkJsonConfig() error {
	if nil == net.ParseIP(common.JsonConfigs.ClientBindIpAddress).To4() &&
		nil == net.ParseIP(common.JsonConfigs.ClientBindIpAddress).To16() {
		if "127.0.0.1" != common.JsonConfigs.ClientBindIpAddress &&
			"0.0.0.0" != common.JsonConfigs.ClientBindIpAddress {
			return fmt.Errorf("common.JsonConfigs.ClientBindIpAddress[%v] is invalid ipv4 address in the config.json file", common.JsonConfigs.ClientBindIpAddress)
		}
	}

	if nil == net.ParseIP(common.JsonConfigs.ClientSendToIpv4Address).To4() {
		if "127.0.0.1" != common.JsonConfigs.ClientBindIpAddress &&
			"0.0.0.0" != common.JsonConfigs.ClientBindIpAddress {
			return fmt.Errorf("common.JsonConfigs.ClientSendToIpv4Address[%v] is invalid ipv4 address in the config.json file", common.JsonConfigs.ClientSendToIpv4Address)
		}
	}

	if nil == net.ParseIP(common.JsonConfigs.ClientSendToIpv6Address).To16() {
		if "::1" != common.JsonConfigs.ClientBindIpAddress {
			return fmt.Errorf("common.JsonConfigs.ClientSendToIpv6Address[%v] is invalid ipv6 address in the config.json file", common.JsonConfigs.ClientSendToIpv6Address)
		}
	}

	return nil
}

func checkFlags() error {
	if 0 != len(common.FlagInfos.ClientBindIpAddress) &&
		nil == net.ParseIP(common.FlagInfos.ClientBindIpAddress) {
		return fmt.Errorf("common.FlagInfos.ClientBindIpAddress[%v] is invalid address, please check -b option", common.FlagInfos.ClientBindIpAddress)
	}

	if 0 == len(common.FlagInfos.ClientBindIpAddress) {
		common.FlagInfos.ClientBindIpAddress = common.JsonConfigs.ClientBindIpAddress
	}

	if false == strings.Contains(common.FlagInfos.ClientBindIpAddress, ":") {
		sendToServerIpAddress = common.JsonConfigs.ClientSendToIpv4Address
	} else {
		sendToServerIpAddress = common.JsonConfigs.ClientSendToIpv6Address
	}

	return nil
}

func parsePort() error {
	if !common.FlagInfos.UsingTcp &&
		!common.FlagInfos.UsingUdp &&
		!common.FlagInfos.UsingHttp &&
		!common.FlagInfos.UsingHttps &&
		!common.FlagInfos.UsingQuic &&
		!common.FlagInfos.UsingDns &&
		0 == common.FlagInfos.SentToServerPort {
		common.FlagInfos.UsingHttp = true
		common.Warn("Please use one of options: -tcp, -udp, -http, -https, -quic, -dns, -dport, default using Http protocol.")
	}

	if 0 != common.FlagInfos.SentToServerPort {
		for _, port := range common.JsonConfigs.ServerTcpListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingTcp = true
				return nil
			}
		}

		for _, port := range common.JsonConfigs.ServerUdpListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingUdp = true
				return nil
			}
		}

		for _, port := range common.JsonConfigs.ServerHttpListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingHttp = true
				return nil
			}
		}

		for _, port := range common.JsonConfigs.ServerHttpsListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingHttps = true
				return nil
			}
		}

		for _, port := range common.JsonConfigs.ServerQuicListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingQuic = true
				return nil
			}
		}

		for _, port := range common.JsonConfigs.ServerDnsListenPorts {
			if port == common.FlagInfos.SentToServerPort {
				common.FlagInfos.UsingDns = true
				return nil
			}
		}

		return fmt.Errorf("Please specify a correct destination port using -dport")
	}

	if common.FlagInfos.UsingTcp {
		if 0 == len(common.JsonConfigs.ServerTcpListenPorts) {
			return fmt.Errorf("Please configure the [ServerTcpListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerTcpListenPorts[0]
		return nil
	}

	if common.FlagInfos.UsingUdp {
		if 0 == len(common.JsonConfigs.ServerUdpListenPorts) {
			return fmt.Errorf("Please configure the [ServerUdpListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerUdpListenPorts[0]
		return nil
	}

	if common.FlagInfos.UsingHttp {
		if 0 == len(common.JsonConfigs.ServerHttpListenPorts) {
			return fmt.Errorf("Please configure the [ServerHttpListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerHttpListenPorts[0]
		return nil
	}

	if common.FlagInfos.UsingHttps {
		if 0 == len(common.JsonConfigs.ServerHttpsListenPorts) {
			return fmt.Errorf("Please configure the [ServerHttpsListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerHttpsListenPorts[0]
		return nil
	}

	if common.FlagInfos.UsingQuic {
		if 0 == len(common.JsonConfigs.ServerQuicListenPorts) {
			return fmt.Errorf("Please configure the [ServerQuicListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerQuicListenPorts[0]
		return nil
	}

	if common.FlagInfos.UsingDns {
		if 0 == len(common.JsonConfigs.ServerDnsListenPorts) {
			return fmt.Errorf("Please configure the [ServerDnsListenPorts] in the config.json file")
		}

		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerDnsListenPorts[0]
		return nil
	}

	return nil
}
