package main

import (
	"../common"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"net"
)

func sendByIEEEQuic() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(common.Configs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(sentToServerPort)}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: localAddr.IP, Port: localAddr.Port})
	if err != nil {
		panic(err)
	}

	session, err := quic.Dial(udpConn, remoteAddr, remoteAddr.String(), &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"ieee-quic"}}, nil)
	if err != nil {
		panic(fmt.Sprintf("IEEE Quic client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}
	defer session.Close()

	stream, err := session.OpenStreamSync(context.Background())
	defer stream.Close()
	if err != nil {
		panic(fmt.Sprintf("IEEE Quic client[%v]----Quic server[%v] open stream failed, err : %v\n", session.LocalAddr(), session.RemoteAddr(), err.Error()))
		return
	}

	fmt.Printf("IEEE Quic client bind on %v, will sent data to %v\n", common.Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= clientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(common.Configs.ClientSendData))
		if err != nil {
			fmt.Printf("IEEE Quic client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			fmt.Printf("IEEE Quic client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("IEEE Quic client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", session.LocalAddr(), session.RemoteAddr(), i, common.Configs.ClientSendData, recvBuffer[:n])
	}
}
