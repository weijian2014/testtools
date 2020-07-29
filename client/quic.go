package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"testtools/common"
	"time"

	"github.com/lucas-clemente/quic-go"
)

func sendByQuic(localAddr, remoteAddr *common.IpAndPort) {
	lAddr, err := net.ResolveUDPAddr("udp", localAddr.String())
	if nil != err {
		panic(err)
	}

	rAddr, err := net.ResolveUDPAddr("udp", remoteAddr.String())
	if nil != err {
		panic(err)
	}

	udpConn, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		panic(err)
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	session, err := quic.Dial(udpConn, rAddr, rAddr.String(), tlsConf, nil)
	if err != nil {
		panic(fmt.Sprintf("Quic client dial with %v failed, err : %v\n", remoteAddr.String(), err.Error()))
	}

	stream, err := session.OpenStreamSync(context.Background())
	defer stream.Close()
	if err != nil {
		panic(fmt.Sprintf("Quic client[%v]----Quic server[%v] open stream failed, err : %v\n", session.LocalAddr(), session.RemoteAddr(), err.Error()))
	}

	common.Info("Quic client bind on %v, will sent data to %v\n", localAddr.String(), remoteAddr.String())
	if 0 != common.FlagInfos.WaitingSeconds {
		common.Info("Quic client waiting %v...\n", common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			common.Warn("Quic client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			common.Warn("Quic client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		common.Info("Quic client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", session.LocalAddr(), session.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
