package server

import (
	"errors"
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

	// 创建./files/目录
	_, err = os.Stat(uploadPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(uploadPath, os.ModePerm)
		if nil != err {
			panic(err)
		}
	}

	// 在./files/目录下创建一个test.txt文件， 并写入ServerSendData数据
	testFile, err := os.Create(uploadPath + "test.txt")
	if nil != err {
		panic(err)
	}
	testFile.Write([]byte(common.JsonConfigs.ServerSendData))
	testFile.Write([]byte("\n"))
	testFile.Close()

	listenAddr := &common.IpAndPort{Ip: common.JsonConfigs.ServerListenHost, Port: 0}

	// Start Tcp server
	for _, port := range common.JsonConfigs.ServerTcpListenPorts {
		listenAddr.Port = port
		go startTcpServer(fmt.Sprintf("TcpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}
	time.Sleep(time.Duration(50) * time.Millisecond)

	// Start Udp server
	for _, port := range common.JsonConfigs.ServerUdpListenPorts {
		listenAddr.Port = port
		go startUdpServer(fmt.Sprintf("UdpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}
	time.Sleep(time.Duration(50) * time.Millisecond)

	// Start Http server
	for _, port := range common.JsonConfigs.ServerHttpListenPorts {
		listenAddr.Port = port
		go startHttpServer(fmt.Sprintf("HttpServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}
	time.Sleep(time.Duration(50) * time.Millisecond)

	// Start Https server
	for i, port := range common.JsonConfigs.ServerHttpsListenPorts {
		listenAddr.Port = port
		go startHttpsServer(fmt.Sprintf("HttpsServer-%v", port), *listenAddr)
		if 0 == i {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	time.Sleep(time.Duration(400) * time.Millisecond)

	// Start Quic server
	for i, port := range common.JsonConfigs.ServerQuicListenPorts {
		listenAddr.Port = uint16(port)
		go startQuicServer(fmt.Sprintf("QuicServer-%v", port), listenAddr)
		if 0 == i {
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
	time.Sleep(time.Duration(400) * time.Millisecond)

	// Start Dns server
	if 0 != len(common.JsonConfigs.ServerTcpListenPorts) {
		saveDnsEntrys()
	}
	for _, port := range common.JsonConfigs.ServerDnsListenPorts {
		listenAddr.Port = port
		go startDnsServer(fmt.Sprintf("DnsServer-%v", port), listenAddr)
		time.Sleep(time.Duration(5) * time.Millisecond)
	}
	time.Sleep(time.Duration(50) * time.Millisecond)

	if 0 == len(common.JsonConfigs.ServerHttpListenPorts) {
		HttpServerGuide(80)
	} else {
		HttpServerGuide(common.JsonConfigs.ServerHttpListenPorts[0])
	}

	if 0 == len(common.JsonConfigs.ServerHttpsListenPorts) {
		HttpsServerGuide(443)
	} else {
		HttpsServerGuide(common.JsonConfigs.ServerHttpsListenPorts[0])
	}

	printDnsServerEntrys()
	common.System("\nJson config: %+v\n\n", common.JsonConfigs)

	var sleepInterval uint64 = 60 * 60
	for {
		time.Sleep(time.Duration(sleepInterval) * time.Second)
		common.System("Service Statistics(interval %v second):\n\tTCP: %v\n\tUDP: %v\n\tHTTP: %v\n\tHTTPS: %v\n\tQUIC: %v\n\tDNS: %v",
			sleepInterval, serverTcpCount, serverUdpCount, serverHttpCount, serverHttpsCount, serverQuicCount, serverDnsCount)
	}
}

func checkJsonConfig() error {
	// Tcp
	for _, port := range common.JsonConfigs.ServerTcpListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	// Udp
	for _, port := range common.JsonConfigs.ServerUdpListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	// Http
	for _, port := range common.JsonConfigs.ServerHttpListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	// Https
	for _, port := range common.JsonConfigs.ServerHttpsListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	// Quic
	for _, port := range common.JsonConfigs.ServerQuicListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	// Dns
	for _, port := range common.JsonConfigs.ServerDnsListenPorts {
		if 0 > port || 65535 < port {
			return errors.New(fmt.Sprintf("Listen port[%v] invalid of config.json file", port))
		}
	}

	return nil
}
