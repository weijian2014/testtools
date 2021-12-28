package server

import (
	"fmt"
	"net"
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
		_, port, err := common.GetIpAndPort(common.JsonConfigs.ServerHttpListenHosts[0])
		if nil != err {
			panic(err)
		}
		HttpServerGuide(port)
	}
	common.Debug("Show the http server guide done!\n")

	if 0 == len(common.JsonConfigs.ServerHttpsListenHosts) {
		HttpsServerGuide(443)
	} else {
		_, port, err := common.GetIpAndPort(common.JsonConfigs.ServerHttpsListenHosts[0])
		if nil != err {
			panic(err)
		}
		HttpsServerGuide(port)
	}
	common.Debug("Show the https server guide done!\n")

	printDnsServerEntrys()
	common.System("\nJson config:%+v\n\n", common.JsonConfigs)
	go startConfigFileWatcher()
	common.Debug("Start the config file watcher done!\n")
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
	listenAddr := common.IpAndPort{Ip: "0.0.0.0", Port: 0}

	// Init Tcp server
	for _, host := range common.JsonConfigs.ServerTcpListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initTcpServer(fmt.Sprintf("TcpServer-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	// Init Udp server
	for _, host := range common.JsonConfigs.ServerUdpListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initUdpServer(fmt.Sprintf("UdpServer-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	// Init Quic server
	for _, host := range common.JsonConfigs.ServerQuicListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initQuicServer(fmt.Sprintf("QuicServer-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(150) * time.Millisecond)
	}

	// Init Http server
	for _, host := range common.JsonConfigs.ServerHttpListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initHttpServer(fmt.Sprintf("HttpServer-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}

	if 0 != len(common.JsonConfigs.ServerHttpsListenHosts) ||
		0 != len(common.JsonConfigs.ServerHttp2ListenHosts) {
		err := common.Mkdir(certificatePath)
		if nil != err {
			panic(err)
		}
	}

	// Init Https server
	if 0 != len(common.JsonConfigs.ServerHttpsListenHosts) {
		prepareHttpsCert()
	}
	for _, host := range common.JsonConfigs.ServerHttpsListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initHttpsServer(fmt.Sprintf("HttpsServer-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(150) * time.Millisecond)
	}

	// Init Http2 server
	if 0 != len(common.JsonConfigs.ServerHttp2ListenHosts) {
		prepareHttp2Cert()
	}
	for _, host := range common.JsonConfigs.ServerHttp2ListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initHttp2Server(fmt.Sprintf("Http2Server-%v", listenAddr.Port), listenAddr)
		time.Sleep(time.Duration(150) * time.Millisecond)
	}

	// Init Dns server
	if 0 != len(common.JsonConfigs.ServerDnsListenHosts) {
		saveDnsEntrys()
	}
	for _, host := range common.JsonConfigs.ServerDnsListenHosts {
		listenAddr.Ip, listenAddr.Port, _ = common.GetIpAndPort(host)
		initDnsServer(fmt.Sprintf("DnsServer-%v", listenAddr.Port), listenAddr)
	}
}

func checkJsonConfig() error {
	// Tcp
	for _, host := range common.JsonConfigs.ServerTcpListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for tcp server", host)
		}
	}

	// Udp
	for _, host := range common.JsonConfigs.ServerUdpListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for udp server", host)
		}
	}

	// Quic
	for _, host := range common.JsonConfigs.ServerQuicListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for quic server", host)
		}
	}

	// Http
	for _, host := range common.JsonConfigs.ServerHttpListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for http server", host)
		}
	}

	// Https
	for _, host := range common.JsonConfigs.ServerHttpsListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for https server", host)
		}
	}

	// Dns
	for _, host := range common.JsonConfigs.ServerDnsListenHosts {
		ip, _, err := common.GetIpAndPort(host)
		if nil != err {
			return err
		}

		if nil == net.ParseIP(ip).To4() &&
			nil == net.ParseIP(ip).To16() {
			return fmt.Errorf("Listen host[%v] invalid of config.json file for dns server", host)
		}
	}

	return nil
}
