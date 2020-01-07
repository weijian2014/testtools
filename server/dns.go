package server

import (
	"errors"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/url"
	"regexp"
	"strings"
	"testtools/common"
)

func startDnsServer(serverName string) {
	saveDnsEntrys()

	serverAddress := fmt.Sprintf("%v:%v", common.JsonConfigs.ServerListenHost, common.JsonConfigs.ServerDnsListenPort)
	udp, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v server startup, listen on %v\n", serverName, serverAddress)

	for {
		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			fmt.Printf("%v server[%v]----Dns client[%v] receive failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		var requestMessage dnsmessage.Message
		err = requestMessage.Unpack(recvBuffer)
		if nil != err {
			fmt.Printf("%v server[%v]----Dns client[%v] unpack failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		//fmt.Printf("Dns server[%v]----Dns client[%v], recv msg:\n\t%+v\n", conn.LocalAddr(), remoteAddress, requestMessage)
		questionCount := len(requestMessage.Questions)
		if 0 == questionCount {
			fmt.Printf("%v server[%v]----Dns client[%v] question count is zero\n", serverName, conn.LocalAddr(), remoteAddress)
			continue
		} else {
			requestMessage.Header.Response = true
			requestMessage.Header.Authoritative = true
		}

		var answers []dnsmessage.Resource
		var tmp string
		for _, question := range requestMessage.Questions {
			h := dnsmessage.ResourceHeader{
				Name:  question.Name,
				Type:  question.Type,
				Class: question.Class,
				TTL:   3600,
			}

			if dnsmessage.TypeA == question.Type {
				ipv4, isOk := dnsAEntrys[question.Name.String()]
				if isOk {
					ip := net.ParseIP(ipv4).To4()
					b := &dnsmessage.AResource{
						A: [4]byte{ip[0], ip[1], ip[2], ip[3]},
					}
					answers = append(answers, dnsmessage.Resource{Header: h, Body: b})
					tmp += ipv4
					tmp += ", "
				}
			} else if dnsmessage.TypeAAAA == question.Type {
				ipv6, isOk := dns4AEntrys[question.Name.String()]
				if isOk {
					ip := net.ParseIP(ipv6).To16()
					b := &dnsmessage.AAAAResource{
						AAAA: [16]byte{
							byte(ip[0]), byte(ip[1]), byte(ip[2]), byte(ip[3]),
							byte(ip[4]), byte(ip[5]), byte(ip[6]), byte(ip[7]),
							byte(ip[8]), byte(ip[9]), byte(ip[10]), byte(ip[11]),
							byte(ip[12]), byte(ip[13]), byte(ip[14]), byte(ip[15]),
						},
					}
					answers = append(answers, dnsmessage.Resource{Header: h, Body: b})
					tmp += ipv6
					tmp += ", "
				}
			} else {
				//fmt.Printf("%v server[%v]----Dns client[%v] question[%d] is not A or AAAA\n", serverName, conn.LocalAddr(), remoteAddress, i+1)
				continue
			}

			requestMessage.Answers = answers
		}

		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			fmt.Printf("%v server[%v]----Dns client[%v] pack failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}
		_, err = conn.WriteToUDP(packed, remoteAddress)
		if err != nil {
			fmt.Printf("%v server[%v]----Dns client[%v] send failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		tmp = strings.TrimRight(tmp, ", ")

		serverDnsTimes++
		fmt.Printf("%v server[%v]----Dns client[%v]:\n\tquestion: %+v\n\t answers: %+v\n",
			serverName, conn.LocalAddr(), remoteAddress, requestMessage.Questions, tmp)
	}
}

func printDnsServerEntrys() {
	if 0 != len(dnsAEntrys) {
		fmt.Printf("Dns server a record:\n")
	}
	for k, v := range dnsAEntrys {
		fmt.Printf("\t%v ---- %v\n", k, v)
	}

	if 0 != len(dns4AEntrys) {
		fmt.Printf("Dns server aaaa record:\n")
	}
	for k, v := range dns4AEntrys {
		fmt.Printf("\t%v ---- %v\n", k, v)
	}
	fmt.Printf("\n")
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
	aEntryMap := common.JsonConfigs.ServerDnsAEntrys.(map[string]interface{})
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
	aaaaEntryMap := common.JsonConfigs.ServerDns4AEntrys.(map[string]interface{})
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
