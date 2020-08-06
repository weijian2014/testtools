package main

import (
	"flag"
	"fmt"
	"os"
	"testtools/client"
	"testtools/common"
	"testtools/server"
)

var (
	tmpSentToServerPort uint = 0
)

func init() {

	// common option
	flag.BoolVar(&common.FlagInfos.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.FlagInfos.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.BoolVar(&common.FlagInfos.IsServer, "server", false, "As server running, default as client")

	// client option
	flag.StringVar(&common.FlagInfos.ClientBindIpAddress, "s", "", "The source IP address of client\n"+
		"This parameter takes precedence over clientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")

	flag.StringVar(&common.FlagInfos.ClientSendToIpAddress, "d", "", "The destination IP address of client\n"+
		"This parameter takes precedence over ClientSendToIpv4Address or ClientSendToIpv6Address in the config.json file\n")

	flag.BoolVar(&common.FlagInfos.UsingTcp, "tcp", false, "Using TCP protocol")
	flag.BoolVar(&common.FlagInfos.UsingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&common.FlagInfos.UsingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&common.FlagInfos.UsingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&common.FlagInfos.UsingQuic, "quic", false, "Using QUIC protocol")
	flag.BoolVar(&common.FlagInfos.UsingDns, "dns", false, "Using DNS protocol")
	flag.UintVar(&tmpSentToServerPort, "dport", 0, "The port of server, valid only for UDP, TCP, QUIC protocols")
	flag.Uint64Var(&common.FlagInfos.WaitingSeconds, "w", 0, "The second waiting to send before, support TCP, UDP, QUIC and DNS protocol")
	flag.Uint64Var(&common.FlagInfos.ClientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, QUIC protocols")
	flag.Parse()
}

func main() {
	var logLevel int = common.JsonConfigs.CommonLogLevel
	common.LoggerInit(logLevel, common.JsonConfigs.CommonLogRoll, "")

	if common.FlagInfos.IsHelp {
		flag.Usage()
		server.HttpServerGuide(80)
		server.HttpsServerGuide(443)
		common.System("\nJson config: %+v\n\n", common.JsonConfigs)
		return
	}

	common.FlagInfos.SentToServerPort = uint16(tmpSentToServerPort)
	_, err := os.Stat(common.FlagInfos.ConfigFileFullPath)
	if os.IsNotExist(err) {
		common.FlagInfos.ConfigFileFullPath = common.CurrDir + "/config/config.json"
		_, err := os.Stat(common.FlagInfos.ConfigFileFullPath)
		if os.IsNotExist(err) {
			common.FlagInfos.ConfigFileFullPath = common.CurrDir + "/../config/config.json"
		}

		_, err = os.Stat(common.FlagInfos.ConfigFileFullPath)
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("Please using -f option specifying a configuration file"))
		}
	}

	common.JsonConfigs, err = common.LoadConfigFile(common.FlagInfos.ConfigFileFullPath)
	if nil != err {
		panic(err)
	}

	if common.FlagInfos.IsServer {
		server.StartServer()
	} else {
		client.StartClient()
	}
}
