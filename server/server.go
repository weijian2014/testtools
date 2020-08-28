package server

import (
	"fmt"
	"os"
	"testtools/common"
	"time"
)

var (
	uploadPath       = ""
	certificatePath  = ""
	dnsAEntrys       map[string]string
	dns4AEntrys      map[string]string
	serverTcpCount   uint64 = 0
	serverUdpCount   uint64 = 0
	serverHttpCount  uint64 = 0
	serverHttpsCount uint64 = 0
	serverQuicCount  uint64 = 0
	serverDnsCount   uint64 = 0
)

func init() {
	uploadPath = common.CurrDir + "/files/"
	certificatePath = common.CurrDir + "/cert/"
}

func StartServer() {
	err := checkJsonConfig()
	if nil != err {
		panic(err)
	}

	common.Debug("Check the %v file done!\n", common.FlagInfos.ConfigFileFullPath)

	// 创建./files/目录
	_, err = os.Stat(uploadPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(uploadPath, os.ModePerm)
		if nil != err {
			panic(err)
		}
	}

	common.Debug("Create the %v directory done!\n", uploadPath)

	// 在./files/目录下创建一个test.txt文件， 并写入ServerSendData数据
	testFileFullPath := uploadPath + "test.txt"
	testFile, err := os.Create(testFileFullPath)
	if nil != err {
		panic(err)
	}
	testFile.Write([]byte(common.JsonConfigs.ServerSendData))
	testFile.Write([]byte("\n"))
	testFile.Close()

	common.Debug("Create the %v file done!\n", testFileFullPath)

	initAllServer()
	common.Debug("Initialize all servers done!\n")
	time.Sleep(time.Duration(50) * time.Millisecond)

	/// Start all server
	err = startAllServers()
	if nil != err {
		panic(err)
	}

	common.Debug("Start all servers done!\n")
	time.Sleep(time.Duration(200) * time.Millisecond)

	if 0 == len(common.JsonConfigs.ServerHttpListenHosts) {
		HttpServerGuide(80)
	} else {
		ip, port, err := common.GetIpAndPort(common.JsonConfigs.ServerHttpListenHosts[0])
		if nil != err {
			panic(err)
		}
		HttpServerGuide(port)
	}
	common.Debug("Show the http server guide done!\n")

	if 0 == len(common.JsonConfigs.ServerHttpsListenHosts) {
		HttpsServerGuide(443)
	} else {
		ip, port, err := common.GetIpAndPort(common.JsonConfigs.ServerHttpsListenHosts[0])
		if nil != err {
			panic(err)
		}
		HttpsServerGuide(port)
	}
	common.Debug("Show the https server guide done!\n")

	printDnsServerEntrys()
	common.System("\nJson config:%+v\n\n", common.JsonConfigs)
	go startConfigFileWatcher()
	common.Debug("Start the config file watcher donw!\n")
	time.Sleep(time.Duration(200) * time.Millisecond)
	common.System("All server start ok\n")
	common.System("================================================================================\n")

	for {
		if 0 == common.JsonConfigs.ServerCounterOutputIntervalSeconds {
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			time.Sleep(time.Duration(common.JsonConfigs.ServerCounterOutputIntervalSeconds) * time.Second)
			common.System("Service Statistics(interval %v second):\n\tTCP: %v\n\tUDP: %v\n\tHTTP: %v\n\tHTTPS: %v\n\tQUIC: %v\n\tDNS: %v",
				common.JsonConfigs.ServerCounterOutputIntervalSeconds, serverTcpCount, serverUdpCount,
				serverHttpCount, serverHttpsCount, serverQuicCount, serverDnsCount)
		}
	}
}

func initAllServer() {
	listenAddr := common.IpAndPort{Ip: common.JsonConfigs.ServerListenHost, Port: 0}

	// Init Tcp server
	for _, port := range common.JsonConfigs.ServerTcpListenPorts {
		listenAddr.Port = port
		initTcpServer(fmt.Sprintf("TcpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	// Init Udp server
	for _, port := range common.JsonConfigs.ServerUdpListenPorts {
		listenAddr.Port = port
		initUdpServer(fmt.Sprintf("UdpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	// // Init Special Udp server
	// for _, host := range common.JsonConfigs.ServerUdpListenHosts {
	// 	index := strings.LastIndex(host, ":")
	// 	ip := host[0:index]
	// 	p, err := strconv.ParseUint(host[index+1:], 10, 16)
	// 	if nil != err {
	// 		panic(err)
	// 	}
	// 	port := uint16(p)
	// 	la := common.IpAndPort{Ip: ip, Port: port}
	// 	initSpecialUdpServer(fmt.Sprintf("SpecialUdpServer-%v", port), la)
	// 	time.Sleep(time.Duration(5) * time.Millisecond)
	// }

	// Init Http server
	for _, port := range common.JsonConfigs.ServerHttpListenPorts {
		listenAddr.Port = port
		initHttpServer(fmt.Sprintf("HttpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	// Init Https server
	if 0 != len(common.JsonConfigs.ServerHttpsListenPorts) {
		prepareCert()
	}
	for _, port := range common.JsonConfigs.ServerHttpsListenPorts {
		listenAddr.Port = port
		initHttpsServer(fmt.Sprintf("HttpsServer-%v", port), listenAddr)
		time.Sleep(time.Duration(150) * time.Millisecond)
	}

	// Init Quic server
	for _, port := range common.JsonConfigs.ServerQuicListenPorts {
		listenAddr.Port = uint16(port)
		initQuicServer(fmt.Sprintf("QuicServer-%v", port), listenAddr)
		time.Sleep(time.Duration(150) * time.Millisecond)
	}

	// Init Dns server
	if 0 != len(common.JsonConfigs.ServerDnsListenPorts) {
		saveDnsEntrys()
	}
	for _, port := range common.JsonConfigs.ServerDnsListenPorts {
		listenAddr.Port = port
		initDnsServer(fmt.Sprintf("DnsServer-%v", port), listenAddr)
	}
}

func checkJsonConfig() error {
	// Tcp
	for _, port := range common.JsonConfigs.ServerTcpListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	// Udp
	for _, port := range common.JsonConfigs.ServerUdpListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	// Http
	for _, port := range common.JsonConfigs.ServerHttpListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	// Https
	for _, port := range common.JsonConfigs.ServerHttpsListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	// Quic
	for _, port := range common.JsonConfigs.ServerQuicListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	// Dns
	for _, port := range common.JsonConfigs.ServerDnsListenPorts {
		if 0 > port || 65535 < port {
			return fmt.Errorf("Listen port[%v] invalid of config.json file", port)
		}
	}

	return nil
}
