package main

import (
	"../common"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
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
		printHttpServerGuide()
		printHttpsServerGuide()
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

	go startTcpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startUdpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpsServer()
	time.Sleep(time.Duration(200) * time.Millisecond)

	go startDnsServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startIeeeQuicServer()

	time.Sleep(time.Duration(200) * time.Millisecond)
	printHttpServerGuide()
	printHttpsServerGuide()
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

	if 0 > common.Configs.ServerTcpListenPort || 65535 < common.Configs.ServerTcpListenPort {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", common.Configs.ServerTcpListenPort))
	}

	if 0 > common.Configs.ServerUdpListenPort || 65535 < common.Configs.ServerUdpListenPort {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", common.Configs.ServerUdpListenPort))
	}

	if 0 > common.Configs.ServerHttpListenPort || 65535 < common.Configs.ServerHttpListenPort {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", common.Configs.ServerHttpListenPort))
	}

	if 0 > common.Configs.ServerHttpsListenPort || 65535 < common.Configs.ServerHttpsListenPort {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", common.Configs.ServerHttpsListenPort))
	}

	if 0 > common.Configs.ServerIeeeQuicListenPort || 65535 < common.Configs.ServerIeeeQuicListenPort {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", common.Configs.ServerIeeeQuicListenPort))
	}

	if 0 > common.Configs.ServerDnsListenPort || 65535 < common.Configs.ServerDnsListenPort {
		return errors.New(fmt.Sprintf("ServerDnsListenPort[%v] invalid of config.json file", common.Configs.ServerDnsListenPort))
	}

	return nil
}

func checkDomainName(domainName string) error {
	if strings.Contains(domainName, " ") {
		return errors.New(fmt.Sprintf("The domain name %v invalid", domainName))
	}

	if strings.HasPrefix(domainName, "http") {
		return errors.New(fmt.Sprintf("The domain name %v invalid, the prefix has 'http'", domainName))
	}

	if strings.HasPrefix(domainName, "https") {
		return errors.New(fmt.Sprintf("The domain name %v invalid, the prefix has 'https", domainName))
	}

	//支持以http://或者https://开头并且域名中间有/的情况
	isLine := "^((http://)|(https://))?([a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,6}(/)"
	_, err := regexp.MatchString(isLine, domainName)
	if nil != err {
		return err
	}

	//支持以http://或者https://开头并且域名中间没有/的情况
	notLine := "^((http://)|(https://))?([a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,6}"
	_, err = regexp.MatchString(notLine, domainName)
	if nil != err {
		return err
	}

	_, err = url.Parse(domainName)
	if nil != err {
		return err
	}

	return nil
}

func saveDnsEntrys() {
	// 读取配置文件中的A记录到map<domainName, IPv4>
	aEntryMap := common.Configs.ServerDnsAEntrys.(map[string]interface{})
	dnsAEntrys = make(map[string]string, len(aEntryMap)+1)
	for domainName, ip := range aEntryMap {
		if nil != checkDomainName(domainName) {
			panic(fmt.Sprintf("The domain name %v invalid", domainName))
		}

		ipv4 := ip.(string)
		if nil == net.ParseIP(ipv4) ||
			false == strings.Contains(ipv4, ".") {
			panic(fmt.Sprintf("The domain name %v not match valid IPv4 address", domainName))
		}
		dnsAEntrys[domainName+"."] = ipv4
	}
	dnsAEntrys["www.example.com."] = "127.0.0.1"

	// 读取配置文件中的AAAA记录到map<domainName, IPv6>
	aaaaEntryMap := common.Configs.ServerDns4AEntrys.(map[string]interface{})
	dns4AEntrys = make(map[string]string, len(aaaaEntryMap)+1)
	for domainName, ip := range aaaaEntryMap {
		if nil != checkDomainName(domainName) {
			panic(fmt.Sprintf("The domain name %v invalid", domainName))
		}

		ipv6 := ip.(string)
		if nil == net.ParseIP(ipv6) ||
			false == strings.Contains(ipv6, ":") {
			panic(fmt.Sprintf("The domain name %v not match valid IPv6 address", domainName))
		}
		dns4AEntrys[domainName+"."] = ipv6
	}
	dns4AEntrys["www.example.com."] = "::1"
}
