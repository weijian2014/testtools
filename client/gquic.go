package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"testtools/common"
	"time"

	"github.com/lucas-clemente/quic-go"
)

func sendByGQuic(serverName string, localAddr, remoteAddr *common.IpAndPort) {
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

	session, err := quic.Dial(udpConn, rAddr, rAddr.String(), &tls.Config{InsecureSkipVerify: true}, &quic.Config{
		Versions: []quic.VersionNumber{
			quic.VersionGQUIC39,
			quic.VersionGQUIC43,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("%v client dial with %v failed, err : %v\n", serverName, remoteAddr.String(), err.Error()))
	}
	defer session.Close()

	stream, err := session.OpenStreamSync()
	defer stream.Close()
	if err != nil {
		panic(fmt.Sprintf("%v client[%v]----Quic server[%v] open stream failed, err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), err.Error()))
	}

	common.Info("%v client bind on %v, will sent data to %v\n", serverName, localAddr.String(), remoteAddr.String())
	if 0 != common.FlagInfos.WaitingSeconds {
		common.Info("%v client waiting %v...\n", serverName, common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			common.Warn("%v client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			common.Warn("%v client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			continue
		}

		common.Info("%v client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
