package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testtools/common"
	"time"
)

var (
	uploadPath      = ""
	certificatePath = ""
	dnsAEntrys      map[string]string
	dns4AEntrys     map[string]string
)

func init() {
	flag.BoolVar(&common.IsHelp, "h", false, "Show help")
	flag.StringVar(&common.ConfigFileFullPath, "f", common.CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.Parse()

	_, err := os.Stat(common.ConfigFileFullPath)
	if os.IsNotExist(err) {
		common.ConfigFileFullPath = common.CurrDir + "/../config/config.json"
	}

	common.Configs, err = common.LoadConfigFile(common.ConfigFileFullPath)
	if nil != err {
		panic(err)
	}

	uploadPath = common.CurrDir + "/files/"
	certificatePath = common.CurrDir + "/cert/"
}

func main() {
	err := checkConfigFlie()
	if nil != err {
		panic(err)
	}

	if common.IsHelp {
		flag.Usage()
		printHttpServerGuide(common.Configs.ServerHttpListenPort1)
		printHttpsServerGuide(common.Configs.ServerHttpsListenPort1)
		saveDnsEntrys()
		printDnsServerEntrys()
		fmt.Printf("Json config: %+v\n\n", common.Configs)
		return
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
	testFile.Write([]byte(common.Configs.ServerSendData))
	testFile.Write([]byte("\n"))
	testFile.Close()

	go startTcpServer(common.Configs.ServerTcpListenPort1, "Tcp-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startTcpServer(common.Configs.ServerTcpListenPort2, "Tcp-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startUdpServer(common.Configs.ServerUdpListenPort1, "Udp-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startUdpServer(common.Configs.ServerUdpListenPort2, "Udp-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpServer(common.Configs.ServerHttpListenPort1, "Http-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startHttpServer(common.Configs.ServerHttpListenPort2, "Http-2")
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpsServer(common.Configs.ServerHttpsListenPort1, "Https-1")
	time.Sleep(time.Duration(5) * time.Millisecond)
	go startHttpsServer(common.Configs.ServerHttpsListenPort2, "Https-1")
	time.Sleep(time.Duration(200) * time.Millisecond)

	go startQuicServer(common.Configs.ServerGoogleQuicListenPort1, "gQuic")
	time.Sleep(time.Duration(10) * time.Millisecond)
	go startQuicServer(common.Configs.ServerGoogleQuicListenPort2, "gQuic")
	time.Sleep(time.Duration(200) * time.Millisecond)

	go startDnsServer("Dns")
	time.Sleep(time.Duration(50) * time.Millisecond)

	printHttpServerGuide(common.Configs.ServerHttpListenPort1)
	printHttpsServerGuide(common.Configs.ServerHttpsListenPort1)
	printDnsServerEntrys()

	for {
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func checkConfigFlie() error {
	if "" != common.Configs.ServerListenHost &&
		"localhost" != common.Configs.ServerListenHost &&
		"0.0.0.0" != common.Configs.ServerListenHost &&
		"127.0.0.1" != common.Configs.ServerListenHost &&
		"::" != common.Configs.ServerListenHost {
		isLocal, err := common.IsLocalIP(common.Configs.ServerListenHost)
		if nil != err {
			return err
		} else if !isLocal {
			return errors.New(fmt.Sprintf("ServerListenHost[%v] is not local address of config.json file", common.Configs.ServerListenHost))
		}
	}

	if 0 > common.Configs.ServerTcpListenPort1 || 65535 < common.Configs.ServerTcpListenPort1 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", common.Configs.ServerTcpListenPort1))
	}
	if 0 > common.Configs.ServerTcpListenPort2 || 65535 < common.Configs.ServerTcpListenPort2 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", common.Configs.ServerTcpListenPort2))
	}
	if common.Configs.ServerTcpListenPort1 == common.Configs.ServerTcpListenPort2 {
		return errors.New(fmt.Sprintf("ServerTcpListenPort has to be different of config.json file"))
	}

	if 0 > common.Configs.ServerUdpListenPort1 || 65535 < common.Configs.ServerUdpListenPort1 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", common.Configs.ServerUdpListenPort1))
	}
	if 0 > common.Configs.ServerUdpListenPort2 || 65535 < common.Configs.ServerUdpListenPort2 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", common.Configs.ServerUdpListenPort2))
	}
	if common.Configs.ServerUdpListenPort1 == common.Configs.ServerUdpListenPort2 {
		return errors.New(fmt.Sprintf("ServerUdpListenPort has to be different of config.json file"))
	}

	if 0 > common.Configs.ServerHttpListenPort1 || 65535 < common.Configs.ServerHttpListenPort1 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", common.Configs.ServerHttpListenPort1))
	}
	if 0 > common.Configs.ServerHttpListenPort2 || 65535 < common.Configs.ServerHttpListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", common.Configs.ServerHttpListenPort2))
	}
	if common.Configs.ServerHttpListenPort1 == common.Configs.ServerHttpListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpListenPort has to be different of config.json file"))
	}

	if 0 > common.Configs.ServerHttpsListenPort1 || 65535 < common.Configs.ServerHttpsListenPort1 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", common.Configs.ServerHttpsListenPort1))
	}
	if 0 > common.Configs.ServerHttpsListenPort2 || 65535 < common.Configs.ServerHttpsListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", common.Configs.ServerHttpsListenPort2))
	}
	if common.Configs.ServerHttpsListenPort1 == common.Configs.ServerHttpsListenPort2 {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort has to be different of config.json file"))
	}

	if 0 > common.Configs.ServerGoogleQuicListenPort1 || 65535 < common.Configs.ServerGoogleQuicListenPort1 {
		return errors.New(fmt.Sprintf("ServerGoogleQuicListenPort[%v] invalid of config.json file", common.Configs.ServerGoogleQuicListenPort1))
	}
	if 0 > common.Configs.ServerGoogleQuicListenPort2 || 65535 < common.Configs.ServerGoogleQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerGoogleQuicListenPort[%v] invalid of config.json file", common.Configs.ServerGoogleQuicListenPort2))
	}

	if 0 > common.Configs.ServerIeeeQuicListenPort1 || 65535 < common.Configs.ServerIeeeQuicListenPort1 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", common.Configs.ServerIeeeQuicListenPort1))
	}
	if 0 > common.Configs.ServerIeeeQuicListenPort2 || 65535 < common.Configs.ServerIeeeQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", common.Configs.ServerIeeeQuicListenPort2))
	}
	if common.Configs.ServerIeeeQuicListenPort1 == common.Configs.ServerIeeeQuicListenPort2 {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort has to be different of config.json file"))
	}

	if 0 > common.Configs.ServerDnsListenPort || 65535 < common.Configs.ServerDnsListenPort {
		return errors.New(fmt.Sprintf("ServerDnsListenPort[%v] invalid of config.json file", common.Configs.ServerDnsListenPort))
	}

	return nil
}
