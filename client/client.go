package main

import (
	"../common"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	usingTcp              bool
	usingUdp              bool
	usingHttp             bool
	usingHttps            bool
	usingDns              bool
	usingIEEEQuic         bool
	usingIos              bool
	waitingSecond         int
	clientBindIpAddress   = ""
	clientSendNumbers     = 1
	sendToServerIpAddress string
)

func init() {
	flag.BoolVar(&common.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.IntVar(&waitingSecond, "w", 0, "The second waiting to send, only support TCP protocol")
	flag.StringVar(&clientBindIpAddress, "b", "", "The ip address of client bind\n"+
		"This parameter takes precedence over clientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")
	flag.IntVar(&clientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, iQuic and gQuic protocols")
	flag.BoolVar(&usingTcp, "tcp", true, "Using TCP protocol")
	flag.BoolVar(&usingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&usingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&usingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&usingIEEEQuic, "iquic", false, "Using IEEE QUIC protocol")
	flag.BoolVar(&usingDns, "dns", false, "Using DNS protocol")
	flag.BoolVar(&usingIos, "ios", false, "Using IOS characteristic TCP header sends TCP packets")
	flag.Parse()

	_, err := os.Stat(common.ConfigFileFullPath)
	if os.IsNotExist(err) {
		common.ConfigFileFullPath = common.CurrDir + "/../config/config.json"
	}

	common.Configs, err = common.LoadConfigFile(common.ConfigFileFullPath)
	if nil != err {
		panic(err)
	}

	if usingUdp || usingHttps || usingHttp || usingDns || usingIEEEQuic || usingIos {
		usingTcp = false
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

	if usingDns {
		sendByDns()
		return
	}

	if usingIEEEQuic {
		sendByIEEEQuic()
		return
	}

	if usingIos {
		sendIosByTcp()
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
