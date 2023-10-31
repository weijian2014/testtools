package client

import (
	"fmt"
	"net"
	"strings"
	"testtools/common"
)

func StartClient() {
	err := checkFlags()
	if nil != err {
		panic(err)
	}

	localAddr := &common.IpAndPort{Ip: common.FlagInfos.ClientScrIp, Port: 0}
	remoteAddr := &common.IpAndPort{Ip: common.FlagInfos.ClientDestIp, Port: common.FlagInfos.ClientDestPort}
	// Tcp
	if common.FlagInfos.ClientUsingTcp {
		sendByTcp(localAddr, remoteAddr)
		return
	}

	// Udp
	if common.FlagInfos.ClientUsingUdp {
		sendByUdp(localAddr, remoteAddr)
		return
	}

	// Http
	if common.FlagInfos.ClientUsingHttp {
		sendByHttp10(localAddr, remoteAddr)
		return
	}

	// Https
	if common.FlagInfos.ClientUsingHttps {
		sendByHttp11(localAddr, remoteAddr, false)
		return
	}

	// Http2.0
	if common.FlagInfos.ClientUsingHttp2 {
		sendByHttp20(localAddr, remoteAddr)
		return
	}

	// Http3
	if common.FlagInfos.ClientUsingHttp3 {
		sendByHttp30(localAddr, remoteAddr)
		return
	}

	// Quic
	if common.FlagInfos.ClientUsingQuic {
		sendByQuic(localAddr, remoteAddr)
		return
	}

	// Dns
	if common.FlagInfos.ClientUsingDns {
		sendByDns(localAddr, remoteAddr)
		return
	}
}

func checkFlags() error {
	// -s src IP check
	if len(common.FlagInfos.ClientScrIp) == 0 {
		return fmt.Errorf("ClientScrIp is invalid address, please specify a correct source ip using -s")
	}
	if nil == net.ParseIP(common.FlagInfos.ClientScrIp).To4() &&
		nil == net.ParseIP(common.FlagInfos.ClientScrIp).To16() {
		return fmt.Errorf("ClientScrIp[%v] is invalid address, please check -s option", common.FlagInfos.ClientScrIp)
	}

	// -d dest IP check
	if len(common.FlagInfos.ClientDestIp) == 0 {
		return fmt.Errorf("ClientDestIp is invalid address, please specify a correct destination ip using -d")
	}
	if nil == net.ParseIP(common.FlagInfos.ClientDestIp).To4() &&
		nil == net.ParseIP(common.FlagInfos.ClientDestIp).To16() {
		return fmt.Errorf("ClientDestIp[%v] is invalid address, please check -d option", common.FlagInfos.ClientDestIp)
	}

	if strings.Contains(common.FlagInfos.ClientScrIp, ":") && !strings.Contains(common.FlagInfos.ClientDestIp, ":") {
		return fmt.Errorf("ClientScrIp[%v] and common.FlagInfos.ClientDestIp[%v] are invalid address, please check -s or -d",
			common.FlagInfos.ClientScrIp, common.FlagInfos.ClientDestIp)
	} else if strings.Contains(common.FlagInfos.ClientScrIp, ".") && !strings.Contains(common.FlagInfos.ClientDestIp, ".") {
		return fmt.Errorf("common.FlagInfos.ClientScrIp[%v] and common.FlagInfos.ClientDestIp[%v] are invalid address, please check -s or -d",
			common.FlagInfos.ClientScrIp, common.FlagInfos.ClientDestIp)
	}

	// -dport
	if 0 >= common.FlagInfos.ClientDestPort || 65535 < common.FlagInfos.ClientDestPort {
		return fmt.Errorf("ClientDestPort[%v] is invalid, please specify a correct destination port using -dport", common.FlagInfos.ClientDestPort)
	}

	// -debug
	if 0 > common.FlagInfos.ClientLogLevel || 5 < common.FlagInfos.ClientLogLevel {
		return fmt.Errorf("ClientLogLevel[%v] is invalid, please check -debug", common.FlagInfos.ClientDestPort)
	}

	if !common.FlagInfos.ClientUsingTcp &&
		!common.FlagInfos.ClientUsingUdp &&
		!common.FlagInfos.ClientUsingHttp &&
		!common.FlagInfos.ClientUsingHttps &&
		!common.FlagInfos.ClientUsingHttp2 &&
		!common.FlagInfos.ClientUsingHttp3 &&
		!common.FlagInfos.ClientUsingQuic &&
		!common.FlagInfos.ClientUsingDns {
		common.System("the client protocol NO specified, please use one of options: -tcp, -udp, -http, -https, -http2, http3, -quic, -dns, -dport, default using http protocol")
		common.FlagInfos.ClientUsingHttp = true
	}

	return nil
}
