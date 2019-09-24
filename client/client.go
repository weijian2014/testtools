package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"testtools/common"
)

var (
	usingTcp              bool
	usingUdp              bool
	usingHttp             bool
	usingHttps            bool
	usingGoogleQuic       bool
	usingIEEEQuic         bool
	usingDns              bool
	waitingSecond         int
	clientBindIpAddress   = ""
	clientSendNumbers     = 1
	sendToServerIpAddress string
	sentToServerPort      uint16
)

func init() {
	tmpSentToServerPort := 0
	flag.BoolVar(&common.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.IntVar(&waitingSecond, "w", 0, "The second waiting to send before, only support TCP protocol")
	flag.StringVar(&clientBindIpAddress, "b", "", "The ip address of client bind\n"+
		"This parameter takes precedence over clientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")
	flag.IntVar(&clientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, gQuic and iQuic protocols")
	flag.IntVar(&tmpSentToServerPort, "dport", 0, "The port of server, valid only for UDP, TCP, gQuic and iQuic protocols")
	flag.BoolVar(&usingTcp, "tcp", false, "Using TCP protocol")
	flag.BoolVar(&usingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&usingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&usingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&usingGoogleQuic, "gquic", false, "Using Google QUIC protocol")
	flag.BoolVar(&usingIEEEQuic, "iquic", false, "Using IEEE QUIC protocol")
	flag.BoolVar(&usingDns, "dns", false, "Using DNS protocol")
	flag.Parse()

	sentToServerPort = uint16(tmpSentToServerPort)
	_, err := os.Stat(common.ConfigFileFullPath)
	if os.IsNotExist(err) {
		common.ConfigFileFullPath = common.CurrDir + "/../config/config.json"
	}

	common.Configs, err = common.LoadConfigFile(common.ConfigFileFullPath)
	if nil != err {
		panic(err)
	}
}

func main() {
	err := checkConfigFlie()
	if nil != err {
		panic(err)
	}

	if nil != net.ParseIP(clientBindIpAddress) {
		common.Configs.ClientBindIpAddress = clientBindIpAddress
	}

	if common.IsHelp {
		flag.Usage()
		fmt.Printf("\nJson config: %+v\n\n", common.Configs)
		return
	}

	if false == strings.Contains(common.Configs.ClientBindIpAddress, ":") {
		sendToServerIpAddress = common.Configs.ClientSendToIpv4Address
	} else {
		sendToServerIpAddress = common.Configs.ClientSendToIpv6Address
	}

	err = parsePort()
	if nil != err {
		panic(err)
	}

	if usingTcp {
		sendByTcp()
		return
	}

	if usingUdp {
		sendByUdp()
		return
	}

	if usingHttp {
		sendByHttp()
		return
	}

	if usingHttps {
		sendByHttps()
		return
	}

	if usingGoogleQuic {
		sendByGQuic("gQuic")
		return
	}

	if usingIEEEQuic {
		sendByGQuic("iQuic")
		return
	}

	if usingDns {
		sendByDns()
		return
	}
}

func checkConfigFlie() error {
	if 0 != len(clientBindIpAddress) &&
		nil == net.ParseIP(clientBindIpAddress) {
		return errors.New(fmt.Sprintf("clientBindIpAddress[%v] is invalid address, please check -b option", clientBindIpAddress))
	}

	if nil == net.ParseIP(common.Configs.ClientSendToIpv4Address) ||
		false == strings.Contains(common.Configs.ClientSendToIpv4Address, ".") {
		return errors.New(fmt.Sprintf("ClientSendToIpv4Address[%v] is invalid ipv4 address in the config.json file", common.Configs.ClientSendToIpv4Address))
	}

	if nil == net.ParseIP(common.Configs.ClientSendToIpv6Address) ||
		false == strings.Contains(common.Configs.ClientSendToIpv6Address, ":") {
		return errors.New(fmt.Sprintf("ClientSendToIpv6Address[%v] is invalid ipv6 address in the config.json file", common.Configs.ClientSendToIpv6Address))
	}

	return nil
}

func parsePort() error {
	if !usingTcp && !usingUdp && !usingHttp && !usingHttps && !usingGoogleQuic && !usingIEEEQuic && !usingDns {
		if 0 == sentToServerPort {
			return errors.New("Please use a required option: -tcp, -udp, -http, -https, -gquic, -iquic, -dns, -dport")
		} else {
			if sentToServerPort == common.Configs.ServerTcpListenPort1 ||
				sentToServerPort == common.Configs.ServerTcpListenPort2 {
				usingTcp = true
			} else if sentToServerPort == common.Configs.ServerUdpListenPort1 ||
				sentToServerPort == common.Configs.ServerUdpListenPort2 {
				usingUdp = true
			} else if sentToServerPort == common.Configs.ServerHttpListenPort1 ||
				sentToServerPort == common.Configs.ServerHttpListenPort2 {
				usingHttp = true
			} else if sentToServerPort == common.Configs.ServerHttpsListenPort1 ||
				sentToServerPort == common.Configs.ServerHttpsListenPort2 {
				usingHttps = true
			} else if sentToServerPort == common.Configs.ServerGoogleQuicListenPort1 ||
				sentToServerPort == common.Configs.ServerGoogleQuicListenPort2 {
				usingGoogleQuic = true
			} else if sentToServerPort == common.Configs.ServerIeeeQuicListenPort1 ||
				sentToServerPort == common.Configs.ServerIeeeQuicListenPort2 {
				usingIEEEQuic = true
			} else if sentToServerPort == common.Configs.ServerDnsListenPort {
				usingDns = true
			}
		}
		return nil
	}

	if usingTcp &&
		sentToServerPort != common.Configs.ServerTcpListenPort1 &&
		sentToServerPort != common.Configs.ServerTcpListenPort2 {
		sentToServerPort = common.Configs.ServerTcpListenPort1
		return nil
	}

	if usingUdp &&
		sentToServerPort != common.Configs.ServerUdpListenPort1 &&
		sentToServerPort != common.Configs.ServerUdpListenPort2 {
		sentToServerPort = common.Configs.ServerUdpListenPort1
		return nil
	}

	if usingHttp &&
		sentToServerPort != common.Configs.ServerHttpListenPort1 &&
		sentToServerPort != common.Configs.ServerHttpListenPort2 {
		sentToServerPort = common.Configs.ServerHttpListenPort1
		return nil
	}

	if usingHttps &&
		sentToServerPort != common.Configs.ServerHttpsListenPort1 &&
		sentToServerPort != common.Configs.ServerHttpsListenPort2 {
		sentToServerPort = common.Configs.ServerHttpsListenPort1
		return nil
	}

	if usingGoogleQuic &&
		sentToServerPort != common.Configs.ServerGoogleQuicListenPort1 &&
		sentToServerPort != common.Configs.ServerGoogleQuicListenPort2 {
		sentToServerPort = common.Configs.ServerGoogleQuicListenPort1
		return nil
	}

	if usingIEEEQuic &&
		sentToServerPort != common.Configs.ServerIeeeQuicListenPort1 &&
		sentToServerPort != common.Configs.ServerIeeeQuicListenPort2 {
		sentToServerPort = common.Configs.ServerIeeeQuicListenPort1
		return nil
	}

	if usingDns &&
		sentToServerPort != common.Configs.ServerDnsListenPort {
		sentToServerPort = common.Configs.ServerDnsListenPort
		return nil
	}

	// default
	usingHttp = true
	sentToServerPort = common.Configs.ServerHttpListenPort1

	return nil
}
