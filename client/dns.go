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
	lAddr := &net.UDPAddr{IP: net.ParseIP(localAddr.Ip)}
	rAddr := &net.UDPAddr{IP: net.ParseIP(remoteAddr.Ip), Port: int(remoteAddr.Port)}

	conn, err := net.DialUDP("udp", lAddr, rAddr)
	if err != nil {
		panic(fmt.Sprintf("Dns client dial with %v failed, err : %v\n", rAddr.String(), err.Error()))
	}
	defer conn.Close()

	if 0 != common.FlagInfos.ClientTimeoutSeconds {
		err = conn.SetDeadline(time.Now().Add(time.Duration(common.FlagInfos.ClientTimeoutSeconds) * time.Second))
		if err != nil {
			panic(err)
		}
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

	common.Info("Dns client bind on %v, will sent query to %v\n", lAddr.String(), rAddr.String())
	if 0 != common.FlagInfos.ClientWaitingSeconds {
		common.Info("Dns client waiting %v...\n", common.FlagInfos.ClientWaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.ClientWaitingSeconds) * time.Second)
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
		recvBuffer := make([]byte, common.FlagInfos.ClientRecvBufferSizeBytes)
		_, err = conn.Read(recvBuffer)
		if err != nil {
			common.Warn("Dns client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
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
