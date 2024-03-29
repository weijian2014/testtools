package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"testtools/common"
	"time"

	"github.com/quic-go/quic-go"
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
	defer udpConn.Close()

	if common.FlagInfos.ClientTimeoutSeconds != 0 {
		err = udpConn.SetDeadline(time.Now().Add(time.Duration(common.FlagInfos.ClientTimeoutSeconds) * time.Second))
		if err != nil {
			panic(err)
		}
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         common.FlagInfos.ClientQuicAlpn,
		ServerName:         common.FlagInfos.ClientQuicSni,
	}

	session, err := quic.DialAddr(context.Background(), rAddr.String(), tlsConf, nil)
	if err != nil {
		panic(fmt.Sprintf("Quic client dial with %v failed, err : %v\n", remoteAddr.String(), err.Error()))
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Quic client[%v]----Quic server[%v] open stream failed, err : %v\n", session.LocalAddr(), session.RemoteAddr(), err.Error()))
	}
	defer stream.Close()

	common.Info("Quic client bind on %v, will sent data to %v\n", localAddr.String(), remoteAddr.String())
	if common.FlagInfos.ClientWaitingSeconds != 0 {
		common.Info("Quic client waiting %v...\n", common.FlagInfos.ClientWaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.ClientWaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Quic client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		// receive
		recvBuffer := make([]byte, common.FlagInfos.ClientRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			common.Warn("Quic client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		common.Info("Quic client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", session.LocalAddr(), session.RemoteAddr(), i, common.FlagInfos.ClientSendData, recvBuffer[:n])
	}
}
