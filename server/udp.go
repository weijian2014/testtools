package server

import (
	"fmt"
	"net"
	"testtools/common"
)

func startUdpServer(listenPort uint16, serverName string) {
	serverAddress := fmt.Sprintf("%v:%v", common.JsonConfigs.ServerListenHost, listenPort)
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
			fmt.Printf("%v server[%v]----Udp client[%v] receive failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		// send
		n, err := conn.WriteToUDP([]byte(common.JsonConfigs.ServerSendData), remoteAddress)
		if err != nil {
			fmt.Printf("%v server[%v]----Udp client[%v] send failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		fmt.Printf("%v server[%v]----Udp client[%v]:\n\trecv: %s\n\tsend: %s\n", serverName, conn.LocalAddr(), remoteAddress, recvBuffer[:n], common.JsonConfigs.ServerSendData)
	}
}
