package client

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testtools/common"
)

var (
	sendToServerIpAddress    string
	clientBindIpAddressRange []string
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

	if common.FlagInfos.UsingClientBindIpAddressRange {
		err = parseClientBindIpAddressRange()
		if nil != err {
			panic(err)
		}

		common.Error("Using client ip address range to binding, count [%v], range [%v]~[%v]\n",
			len(clientBindIpAddressRange), clientBindIpAddressRange[0], clientBindIpAddressRange[len(clientBindIpAddressRange)-1])
	}

	if common.FlagInfos.UsingTcp {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendTcpByRange()
		} else {
			sendByTcp(common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingUdp {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendUdpByRange()
		} else {
			sendByUdp(common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingHttp {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendHttpByRange()
		} else {
			sendByHttp(common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingHttps {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendHttpsByRange()
		} else {
			sendByHttps(common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingGoogleQuic {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendGQuicByRange()
		} else {
			sendByGQuic("gQuic", common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingIEEEQuic {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendIQuicByRange()
		} else {
			sendByGQuic("iQuic", common.FlagInfos.ClientBindIpAddress)
		}

		return
	}

	if common.FlagInfos.UsingDns {
		if common.FlagInfos.UsingClientBindIpAddressRange {
			sendDnsByRange()
		} else {
			sendByDns(common.FlagInfos.ClientBindIpAddress)
		}

		return
	}
}

func checkJsonConfig() error {
	if nil == net.ParseIP(common.JsonConfigs.ClientBindIpAddress).To4() &&
		nil == net.ParseIP(common.JsonConfigs.ClientBindIpAddress).To16() {
		if "127.0.0.1" != common.JsonConfigs.ClientBindIpAddress &&
			"0.0.0.0" != common.JsonConfigs.ClientBindIpAddress {
			return errors.New(fmt.Sprintf("common.JsonConfigs.ClientBindIpAddress[%v] is invalid ipv4 address in the config.json file", common.JsonConfigs.ClientBindIpAddress))
		}
	}

	if nil == net.ParseIP(common.JsonConfigs.ClientSendToIpv4Address).To4() {
		if "127.0.0.1" != common.JsonConfigs.ClientBindIpAddress &&
			"0.0.0.0" != common.JsonConfigs.ClientBindIpAddress {
			return errors.New(fmt.Sprintf("common.JsonConfigs.ClientSendToIpv4Address[%v] is invalid ipv4 address in the config.json file", common.JsonConfigs.ClientSendToIpv4Address))
		}
	}

	if nil == net.ParseIP(common.JsonConfigs.ClientSendToIpv6Address).To16() {
		if "::1" != common.JsonConfigs.ClientBindIpAddress {
			return errors.New(fmt.Sprintf("common.JsonConfigs.ClientSendToIpv6Address[%v] is invalid ipv6 address in the config.json file", common.JsonConfigs.ClientSendToIpv6Address))
		}
	}

	return nil
}

func parseClientBindIpAddressRange() error {
	if 0 == len(common.JsonConfigs.ClientBindIpAddressRange) {
		return errors.New(fmt.Sprintf("common.JsonConfigs.ClientBindIpAddressRange is invalid"))
	}

	ip, ipnet, err := net.ParseCIDR(common.JsonConfigs.ClientBindIpAddressRange)
	if err != nil {
		return err
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIp(ip) {
		if !net.ParseIP(common.JsonConfigs.ClientSendToIpv4Address).IsMulticast() {
			clientBindIpAddressRange = append(clientBindIpAddressRange, ip.String())
		}
	}

	if 0 == len(clientBindIpAddressRange) {
		return errors.New(fmt.Sprintf("parse common.JsonConfigs.ClientBindIpAddressRange[%v] fail", common.JsonConfigs.ClientBindIpAddressRange))
	}

	return nil
}

func incIp(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func checkFlags() error {
	if 0 != len(common.FlagInfos.ClientBindIpAddress) &&
		nil == net.ParseIP(common.FlagInfos.ClientBindIpAddress) {
		return errors.New(fmt.Sprintf("common.FlagInfos.ClientBindIpAddress[%v] is invalid address, please check -b option", common.FlagInfos.ClientBindIpAddress))
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
		!common.FlagInfos.UsingGoogleQuic &&
		!common.FlagInfos.UsingIEEEQuic &&
		!common.FlagInfos.UsingDns {
		if 0 == common.FlagInfos.SentToServerPort {
			return errors.New("Please use one of options: -tcp, -udp, -http, -https, -gquic, -iquic, -dns, -dport")
		} else {
			if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerTcpListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerTcpListenPort2 {
				common.FlagInfos.UsingTcp = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerUdpListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerUdpListenPort2 {
				common.FlagInfos.UsingUdp = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerHttpListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerHttpListenPort2 {
				common.FlagInfos.UsingHttp = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerHttpsListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerHttpsListenPort2 {
				common.FlagInfos.UsingHttps = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerGoogleQuicListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerGoogleQuicListenPort2 {
				common.FlagInfos.UsingGoogleQuic = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerIeeeQuicListenPort1 ||
				common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerIeeeQuicListenPort2 {
				common.FlagInfos.UsingIEEEQuic = true
			} else if common.FlagInfos.SentToServerPort == common.JsonConfigs.ServerDnsListenPort {
				common.FlagInfos.UsingDns = true
			}
		}
		return nil
	}

	if common.FlagInfos.UsingTcp &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerTcpListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerTcpListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerTcpListenPort1
		return nil
	}

	if common.FlagInfos.UsingUdp &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerUdpListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerUdpListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerUdpListenPort1
		return nil
	}

	if common.FlagInfos.UsingHttp &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerHttpListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerHttpListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerHttpListenPort1
		return nil
	}

	if common.FlagInfos.UsingHttps &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerHttpsListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerHttpsListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerHttpsListenPort1
		return nil
	}

	if common.FlagInfos.UsingGoogleQuic &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerGoogleQuicListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerGoogleQuicListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerGoogleQuicListenPort1
		return nil
	}

	if common.FlagInfos.UsingIEEEQuic &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerIeeeQuicListenPort1 &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerIeeeQuicListenPort2 {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerIeeeQuicListenPort1
		return nil
	}

	if common.FlagInfos.UsingDns &&
		common.FlagInfos.SentToServerPort != common.JsonConfigs.ServerDnsListenPort {
		common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerDnsListenPort
		return nil
	}

	// default
	common.FlagInfos.UsingHttp = true
	common.FlagInfos.SentToServerPort = common.JsonConfigs.ServerHttpListenPort1

	return nil
}
