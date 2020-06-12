package client

import (
	"fmt"
	"net"
	"strings"
	"testtools/common"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func sendByDns(localAddr, remoteAddr *common.IpAndPort) {
	lAddr, err := net.ResolveUDPAddr("udp", localAddr.String())
	if nil != err {
		panic(err)
	}

	rAddr, err := net.ResolveUDPAddr("udp", remoteAddr.String())
	if nil != err {
		panic(err)
	}

	conn, err := net.DialUDP("udp", lAddr, rAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Dns client dial with %v failed, err : %v\n", remoteAddr.String(), err.Error()))
	}

	var questionType dnsmessage.Type
	if false == strings.Contains(localAddr.Ip, ":") {
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

	common.Info("Dns client bind on %v, will sent query to %v\n", localAddr.String(), remoteAddr.String())
	if 0 != common.FlagInfos.WaitingSeconds {
		common.Info("Dns client waiting %v...\n", common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			common.Warn("Dns client[%v]----Dns server[%v] pack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			continue
		}
		_, err = conn.Write(packed)
		if err != nil {
			common.Warn("Dns client[%v]----Dns server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			continue
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		_, err = conn.Read(recvBuffer)
		if err != nil {
			common.Warn("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			continue
		}
		var responseMessage dnsmessage.Message
		err = responseMessage.Unpack(recvBuffer)
		if nil != err {
			common.Warn("Dns client[%v]----Dns server[%v] unpack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			continue
		}

		if dnsmessage.TypeA == questionType {
			ipv4 := responseMessage.Answers[0].Body.GoString()
			common.Info("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\tanswers: %+v\n",
				conn.LocalAddr(), conn.RemoteAddr(), i, requestMessage.Questions[0], ipv4)
		} else {
			ipv6 := responseMessage.Answers[0].Body.GoString()
			common.Info("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\t answers: %+v\n",
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
