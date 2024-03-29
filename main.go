package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testtools/client"
	"testtools/common"
	"testtools/server"
)

func init() {
	var tmpSentToServerPort uint
	var tmpQuicAlpn string

	// common option
	flag.BoolVar(&common.FlagInfos.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.FlagInfos.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.BoolVar(&common.FlagInfos.IsServer, "server", false, "As server running, default as client")

	// client option
	flag.StringVar(&common.FlagInfos.ClientScrIp, "s", "", "The source IP address of client\n"+
		"This parameter takes precedence over clientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")

	flag.StringVar(&common.FlagInfos.ClientDestIp, "d", "", "The destination IP address of client\n"+
		"This parameter takes precedence over ClientSendToIpv4Address or ClientSendToIpv6Address in the config.json file\n")

	flag.UintVar(&tmpSentToServerPort, "dport", 0, "The port of server, valid only for UDP, TCP, QUIC protocols")

	flag.StringVar(&common.FlagInfos.ClientJsonDataFile, "json", "", "The path of json format data file to send")

	flag.BoolVar(&common.FlagInfos.ClientUsingTcp, "tcp", false, "Using TCP protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingHttp2, "http2", false, "Using HTTP2 protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingHttp3, "http3", false, "Using HTTP3 over QUIC protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingQuic, "quic", false, "Using QUIC protocol")
	flag.BoolVar(&common.FlagInfos.ClientUsingDns, "dns", false, "Using DNS protocol")
	flag.StringVar(&tmpQuicAlpn, "alpn", "", "The ALPN of QUIC protocol, which example as \"aaa,bbb,ccc\", \"ietf-quic-v1-echo-example\" is hard code both in server and client")
	flag.StringVar(&common.FlagInfos.ClientQuicSni, "sni", "", "The SNI of QUIC protocol")
	flag.Uint64Var(&common.FlagInfos.ClientWaitingSeconds, "w", 0, "The second waiting to send before, support TCP, UDP, QUIC and DNS protocol")
	flag.Uint64Var(&common.FlagInfos.ClientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, QUIC protocols")
	flag.IntVar(&common.FlagInfos.ClientLogLevel, "debug", 0, "The client log level, 0-Debug, 1-Info, 2-System, 3-Warn, 4-Error, 5-Fatal")
	flag.Uint64Var(&common.FlagInfos.ClientTimeoutSeconds, "t", 0, "The timeout seconds of client read or write")
	flag.Parse()

	common.FlagInfos.ClientDestPort = uint16(tmpSentToServerPort)
	if common.FlagInfos.ClientUsingQuic {
		common.FlagInfos.ClientQuicAlpn = append(common.FlagInfos.ClientQuicAlpn, common.COMMON_QUIC_ALPN)
		if 0 < len(tmpQuicAlpn) {
			common.FlagInfos.ClientQuicAlpn = append(common.FlagInfos.ClientQuicAlpn, strings.Split(tmpQuicAlpn, ",")...)
		}
	}
}

func main() {
	if common.FlagInfos.IsHelp {
		flag.Usage()
		return
	}

	var logLevel int = 0
	var logRoll int = 50000

	if common.FlagInfos.IsServer {
		// server
		_, err := os.Stat(common.FlagInfos.ConfigFileFullPath)
		if os.IsNotExist(err) {
			common.FlagInfos.ConfigFileFullPath = common.CurrDir + "/config/config.json"
			_, err := os.Stat(common.FlagInfos.ConfigFileFullPath)
			if os.IsNotExist(err) {
				common.FlagInfos.ConfigFileFullPath = common.CurrDir + "/../config/config.json"
			}

			_, err = os.Stat(common.FlagInfos.ConfigFileFullPath)
			if os.IsNotExist(err) {
				panic("Please using -f option specifying a configuration file")
			}
		}

		common.JsonConfigs, err = common.LoadConfigFile(common.FlagInfos.ConfigFileFullPath)
		if nil != err {
			panic(err)
		}

		logLevel = common.JsonConfigs.ServerLogLevel
		logRoll = common.JsonConfigs.ServerLogRoll
	} else {
		// client
		logLevel = common.FlagInfos.ClientLogLevel
		if len(common.FlagInfos.ClientJsonDataFile) != 0 {
			jsonData, err := common.ReadJsonFile(common.FlagInfos.ClientJsonDataFile)
			if nil != err {
				panic(fmt.Sprintf("Read json format data file %v failed, error %v", common.FlagInfos.ClientJsonDataFile, err.Error()))
			} else {
				common.FlagInfos.ClientSendData = jsonData
			}
		} else {
			common.FlagInfos.ClientSendData = "Hello Server"
		}
		common.FlagInfos.ClientRecvBufferSizeBytes = 1024
	}

	// The third parameter "" mean the log output to stdout
	common.LoggerInit(logLevel, logRoll, "")

	err := common.SetUdpReceiveBufferSizeByQuic()
	if nil != err {
		common.System("Set UDP receive buffer size by QUIC failed, err: %v\n", err.Error())
	}

	if common.FlagInfos.IsServer {
		server.StartServer()
	} else {
		// common.System("\nFlag info:%+v\n\n", common.FlagInfos)
		client.StartClient()
	}
}
