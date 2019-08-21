package main

import (
	"../common"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"strings"
)

func startDnsServer() {
	saveDnsEntrys()

	serverAddress := fmt.Sprintf("%v:%v", common.Configs.ServerListenHost, common.Configs.ServerDnsListenPort)
	udp, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Dns   server startup, listen on %v\n", serverAddress)

	for {
		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			fmt.Printf("Dns server[%v]----Dns client[%v] receive failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		var requestMessage dnsmessage.Message
		err = requestMessage.Unpack(recvBuffer)
		if nil != err {
			fmt.Printf("Dns server[%v]----Dns client[%v] unpack failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		//fmt.Printf("Dns server[%v]----Dns client[%v], recv msg:\n\t%+v\n", conn.LocalAddr(), remoteAddress, requestMessage)
		questionCount := len(requestMessage.Questions)
		if 0 == questionCount {
			fmt.Printf("Dns server[%v]----Dns client[%v] question count is zero\n", conn.LocalAddr(), remoteAddress)
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
				//fmt.Printf("Dns server[%v]----Dns client[%v] question[%d] is not A or AAAA\n", conn.LocalAddr(), remoteAddress, i+1)
				continue
			}

			requestMessage.Answers = answers
		}

		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			fmt.Printf("Dns server[%v]----Dns client[%v] pack failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}
		_, err = conn.WriteToUDP(packed, remoteAddress)
		if err != nil {
			fmt.Printf("Dns server[%v]----Dns client[%v] send failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		tmp = strings.TrimRight(tmp, ", ")
		fmt.Printf("Dns server[%v]----Dns client[%v]:\n\tquestion: %+v\n\t answers: %+v\n",
			conn.LocalAddr(), remoteAddress, requestMessage.Questions, tmp)
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
