package client

import (
	"fmt"
	"net"
	"testtools/common"
	"time"
)

func sendByTcp() {
	localAddr := &net.TCPAddr{IP: net.ParseIP(common.JsonConfigs.ClientBindIpAddress)}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(common.FlagInfos.SentToServerPort)}
	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Tcp client connect to %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Tcp client bind on %v, will sent data to %v\n", common.JsonConfigs.ClientBindIpAddress, remoteAddr)
	if 0 != common.FlagInfos.WaitingSeconds {
		fmt.Printf("Tcp client waiting %v...\n", common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			fmt.Printf("Tcp client[%v]----Tcp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Tcp client[%v]----Tcp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Tcp client[%v]----Tcp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
