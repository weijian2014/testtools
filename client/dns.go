package client

import (
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"strings"
	"testtools/common"
)

func sendByDns() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(common.JsonConfigs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(common.JsonConfigs.ServerDnsListenPort)}
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Dns client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	var questionType dnsmessage.Type
	if false == strings.Contains(common.JsonConfigs.ClientBindIpAddress, ":") {
		questionType = dnsmessage.TypeA
	} else {
		questionType = dnsmessage.TypeAAAA
	}
	requestMessage := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 8888,
			Response:           false,
			OpCode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: false,
			RCode:              dnsmessage.RCodeSuccess,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  mustNewName("www.example.com."),
				Type:  questionType,
				Class: dnsmessage.ClassINET,
			},
		},
	}

	fmt.Printf("Dns client bind on %v, will sent query to %v\n", common.JsonConfigs.ClientBindIpAddress, remoteAddr)

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			fmt.Printf("Dns client[%v]----Dns server[%v] pack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}
		_, err = conn.Write(packed)
		if err != nil {
			fmt.Printf("Dns client[%v]----Dns server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		_, err = conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}
		var responseMessage dnsmessage.Message
		err = responseMessage.Unpack(recvBuffer)
		if nil != err {
			fmt.Printf("Dns client[%v]----Dns server[%v] unpack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}

		if dnsmessage.TypeA == questionType {
			ipv4 := responseMessage.Answers[0].Body.GoString()
			fmt.Printf("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\tanswers: %+v\n",
				conn.LocalAddr(), conn.RemoteAddr(), i, requestMessage.Questions[0], ipv4)
		} else {
			ipv6 := responseMessage.Answers[0].Body.GoString()
			fmt.Printf("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\t answers: %+v\n",
				conn.LocalAddr(), conn.RemoteAddr(), i, requestMessage.Questions[0], ipv6)
		}
	}
}

func mustNewName(name string) dnsmessage.Name {
	n, err := dnsmessage.NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}
