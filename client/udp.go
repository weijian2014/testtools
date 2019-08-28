package main

import (
	"../common"
	"fmt"
	"net"
)

func sendByUdp() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(common.Configs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(sentToServerPort)}
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Udp client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Udp client bind on %v, will sent data to %v\n", common.Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= clientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Udp client[%v]----Udp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.Configs.ClientSendData, recvBuffer[:n])
	}
}
