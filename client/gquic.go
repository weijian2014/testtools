package client

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"net"
	"testtools/common"
)

func sendByGQuic(serverName string) {
	localAddr := &net.UDPAddr{IP: net.ParseIP(common.JsonConfigs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(common.FlagInfos.SentToServerPort)}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: localAddr.IP, Port: localAddr.Port})
	if err != nil {
		panic(err)
	}

	session, err := quic.Dial(udpConn, remoteAddr, remoteAddr.String(), &tls.Config{InsecureSkipVerify: true}, &quic.Config{
		Versions: []quic.VersionNumber{
			quic.VersionGQUIC39,
			quic.VersionGQUIC43,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("%v client dial with %v failed, err : %v\n", serverName, remoteAddr, err.Error()))
	}
	defer session.Close()

	stream, err := session.OpenStreamSync()
	defer stream.Close()
	if err != nil {
		panic(fmt.Sprintf("%v client[%v]----Quic server[%v] open stream failed, err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), err.Error()))
		return
	}

	fmt.Printf("%v client bind on %v, will sent data to %v\n", serverName, common.JsonConfigs.ClientBindIpAddress, remoteAddr)

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			fmt.Printf("%v client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			fmt.Printf("%v client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("%v client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", serverName, session.LocalAddr(), session.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
