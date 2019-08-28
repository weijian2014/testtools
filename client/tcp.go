package main

import (
	"../common"
	"fmt"
	"net"
	"time"
)

func sendByTcp() {
	localAddr := &net.TCPAddr{IP: net.ParseIP(common.Configs.ClientBindIpAddress)}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(sentToServerPort)}
	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Tcp client connect to %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Tcp client bind on %v, will sent data to %v\n", common.Configs.ClientBindIpAddress, remoteAddr)
	if 0 != waitingSecond {
		fmt.Printf("Tcp client waiting %v...\n", waitingSecond)
		time.Sleep(time.Duration(waitingSecond) * time.Second)
	}

	for i := 1; i <= clientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Tcp client[%v]----Tcp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Tcp client[%v]----Tcp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Tcp client[%v]----Tcp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.Configs.ClientSendData, recvBuffer[:n])
	}
}
