package client

import (
	"fmt"
	"net"
	"testtools/common"
	"time"
)

func sendByUdp() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(common.FlagInfos.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(sendToServerIpAddress), Port: int(common.FlagInfos.SentToServerPort)}
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Udp client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Udp client bind on %v, will sent data to %v\n", common.FlagInfos.ClientBindIpAddress, remoteAddr)
	if 0 != common.FlagInfos.WaitingSeconds {
		fmt.Printf("Udp client waiting %v...\n", common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Udp client[%v]----Udp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
