package main

import (
	"../common"
	"fmt"
	"net"
)

func startUdpServer(listenPort uint16) {
	serverAddress := fmt.Sprintf("%v:%v", common.Configs.ServerListenHost, listenPort)
	udp, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Udp   server startup, listen on %v\n", serverAddress)

	for {
		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			fmt.Printf("Udp server[%v]----Udp client[%v] receive failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		// send
		n, err := conn.WriteToUDP([]byte(common.Configs.ServerSendData), remoteAddress)
		if err != nil {
			fmt.Printf("Udp server[%v]----Udp client[%v] send failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		fmt.Printf("Udp server[%v]----Udp client[%v]:\n\trecv: %s\n\tsend: %s\n", conn.LocalAddr(), remoteAddress, recvBuffer[:n], common.Configs.ServerSendData)
	}
}
