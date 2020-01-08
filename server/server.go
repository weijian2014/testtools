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
	serverTcpTimes   uint64 = 0
	serverUdpTimes   uint64 = 0
	serverHttpTimes  uint64 = 0
	serverHttpsTimes uint64 = 0
	serverGQuicTimes uint64 = 0
	serverIQuicTimes uint64 = 0
	serverDnsTimes   uint64 = 0
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

	go startTcpServer(common.JsonConfigs.ServerTcpListenPort1, "Tcp-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startTcpServer(common.JsonConfigs.ServerTcpListenPort2, "Tcp-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startUdpServer(common.JsonConfigs.ServerUdpListenPort1, "Udp-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startUdpServer(common.JsonConfigs.ServerUdpListenPort2, "Udp-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpServer(common.JsonConfigs.ServerHttpListenPort1, "Http-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startHttpServer(common.JsonConfigs.ServerHttpListenPort2, "Http-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpsServer(common.JsonConfigs.ServerHttpsListenPort1, "Https-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startHttpsServer(common.JsonConfigs.ServerHttpsListenPort2, "Https-1")
	time.Sleep(time.Duration(400) * time.Millisecond)

	go startGQuicServer(common.JsonConfigs.ServerGoogleQuicListenPort1, "gQuic-1")
	time.Sleep(time.Duration(10) * time.Millisecond)
	go startGQuicServer(common.JsonConfigs.ServerGoogleQuicListenPort2, "gQuic-2")
	time.Sleep(time.Duration(400) * time.Millisecond)

	go startDnsServer("Dns")
	time.Sleep(time.Duration(50) * time.Millisecond)

	HttpServerGuide(common.JsonConfigs.ServerHttpListenPort1)
	HttpsServerGuide(common.JsonConfigs.ServerHttpsListenPort1)
	printDnsServerEntrys()
	common.Error("\nJson config: %+v\n\n", common.JsonConfigs)

	var sleepInterval uint64 = 15
	for {
		time.Sleep(time.Duration(sleepInterval) * time.Second)
		common.Error("Service Statistics(interval=%v):\n\tTCP: %v\n\tUDP: %v\n\tHTTP: %v\n\tHTTPS: %v\n\tGQUIC: %v\n\tIQUIC: %v\n\tDNS: %v",
			sleepInterval, serverTcpTimes, serverUdpTimes, serverHttpTimes, serverHttpsTimes, serverGQuicTimes, serverIQuicTimes, serverDnsTimes)
	}
}

func checkJsonConfig() error {
	if "" != common.JsonConfigs.ServerListenHost &&
		"localhost" != common.JsonConfigs.ServerListenHost &&
		"0.0.0.0" != common.JsonConfigs.ServerListenHost &&
		"127.0.0.1" != common.JsonConfigs.ServerListenHost &&
		"::" != common.JsonConfigs.ServerListenHost {
		isLocal, err := common.IsLocalIP(common.JsonConfigs.ServerListenHost)
		if nil != err {
			return err
		} else if !isLocal {
			return errors.New(fmt.Sprintf("ServerListenHost[%v] is not local address of config.json file", common.JsonConfigs.ServerListenHost))
		}
	}

	if 0 > common.JsonConfigs.ServerTcpListenPort1 || 65535 < common.JsonConfigs.ServerTcpListenPort1 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerTcpListenPort1))
	}
	if 0 > common.JsonConfigs.ServerTcpListenPort2 || 65535 < common.JsonConfigs.ServerTcpListenPort2 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerTcpListenPort2))
	}
	if common.JsonConfigs.ServerTcpListenPort1 == common.JsonConfigs.ServerTcpListenPort2 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort has to be different of config.json file"))
	}

	if 0 > common.JsonConfigs.ServerUdpListenPort1 || 65535 < common.JsonConfigs.ServerUdpListenPort1 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerUdpListenPort1))
	}
	if 0 > common.JsonConfigs.ServerUdpListenPort2 || 65535 < common.JsonConfigs.ServerUdpListenPort2 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerUdpListenPort2))
	}
	if common.JsonConfigs.ServerUdpListenPort1 == common.JsonConfigs.ServerUdpListenPort2 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort has to be different of config.json file"))
	}

	if 0 > common.JsonConfigs.ServerHttpListenPort1 || 65535 < common.JsonConfigs.ServerHttpListenPort1 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerHttpListenPort1))
	}
	if 0 > common.JsonConfigs.ServerHttpListenPort2 || 65535 < common.JsonConfigs.ServerHttpListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerHttpListenPort2))
	}
	if common.JsonConfigs.ServerHttpListenPort1 == common.JsonConfigs.ServerHttpListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort has to be different of config.json file"))
	}

	if 0 > common.JsonConfigs.ServerHttpsListenPort1 || 65535 < common.JsonConfigs.ServerHttpsListenPort1 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerHttpsListenPort1))
	}
	if 0 > common.JsonConfigs.ServerHttpsListenPort2 || 65535 < common.JsonConfigs.ServerHttpsListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerHttpsListenPort2))
	}
	if common.JsonConfigs.ServerHttpsListenPort1 == common.JsonConfigs.ServerHttpsListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort has to be different of config.json file"))
	}

	if 0 > common.JsonConfigs.ServerGoogleQuicListenPort1 || 65535 < common.JsonConfigs.ServerGoogleQuicListenPort1 {
		return errors.New(fmt.Sprintf("ServerGoogleQuicListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerGoogleQuicListenPort1))
	}
	if 0 > common.JsonConfigs.ServerGoogleQuicListenPort2 || 65535 < common.JsonConfigs.ServerGoogleQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerGoogleQuicListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerGoogleQuicListenPort2))
	}

	if 0 > common.JsonConfigs.ServerIeeeQuicListenPort1 || 65535 < common.JsonConfigs.ServerIeeeQuicListenPort1 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerIeeeQuicListenPort1))
	}
	if 0 > common.JsonConfigs.ServerIeeeQuicListenPort2 || 65535 < common.JsonConfigs.ServerIeeeQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerIeeeQuicListenPort2))
	}
	if common.JsonConfigs.ServerIeeeQuicListenPort1 == common.JsonConfigs.ServerIeeeQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort has to be different of config.json file"))
	}

	if 0 > common.JsonConfigs.ServerDnsListenPort || 65535 < common.JsonConfigs.ServerDnsListenPort {
		return errors.New(fmt.Sprintf("ServerDnsListenPort[%v] invalid of config.json file", common.JsonConfigs.ServerDnsListenPort))
	}

	return nil
}
