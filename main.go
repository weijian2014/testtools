package main

import (
	"flag"
	"fmt"
	"os"
	"testtools/client"
	"testtools/common"
	"testtools/server"
)

func init() {
	var tmpSentToServerPort uint

	// common option
	flag.BoolVar(&common.FlagInfos.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.FlagInfos.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.BoolVar(&common.FlagInfos.IsServer, "server", false, "As server running, default as client")

	// client option
	flag.StringVar(&common.FlagInfos.ClientBindIpAddress, "b", "", "The ip address of client bind\n"+
		"This parameter takes precedence over clientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")
	flag.BoolVar(&common.FlagInfos.UsingTcp, "tcp", false, "Using TCP protocol")
	flag.BoolVar(&common.FlagInfos.UsingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&common.FlagInfos.UsingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&common.FlagInfos.UsingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&common.FlagInfos.UsingGoogleQuic, "gquic", false, "Using Google QUIC protocol")
	flag.BoolVar(&common.FlagInfos.UsingIEEEQuic, "iquic", false, "Using IEEE QUIC protocol, unavailable")
	flag.BoolVar(&common.FlagInfos.UsingDns, "dns", false, "Using DNS protocol")
	flag.UintVar(&tmpSentToServerPort, "dport", 0, "The port of server, valid only for UDP, TCP, gQuic and iQuic protocols")
	flag.Uint64Var(&common.FlagInfos.WaitingSeconds, "w", 0, "The second waiting to send before, support TCP, UDP, gQuic and DNS protocol")
	flag.Uint64Var(&common.FlagInfos.ClientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, gQuic and iQuic protocols")
	flag.Parse()

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
}

func main() {
	if common.FlagInfos.IsHelp {
		flag.Usage()
		server.HttpServerGuide(common.JsonConfigs.ServerHttpListenPort1)
		server.HttpsServerGuide(common.JsonConfigs.ServerHttpsListenPort1)
		fmt.Printf("\nJson config: %+v\n\n", common.JsonConfigs)
		return
	}

	if common.FlagInfos.IsServer {
		server.StartServer()
	} else {
		client.StartClient()
	}
}
