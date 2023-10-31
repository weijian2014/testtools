package client

import (
	"fmt"
	"net"
	"testtools/common"
	"time"
)

func sendByUdp(localAddr, remoteAddr *common.IpAndPort) {
	lAddr, err := net.ResolveUDPAddr("udp", localAddr.String())
	if nil != err {
		panic(err)
	}

	rAddr, err := net.ResolveUDPAddr("udp", remoteAddr.String())
	if nil != err {
		panic(err)
	}

	conn, err := net.DialUDP("udp", lAddr, rAddr)
	if err != nil {
		panic(fmt.Sprintf("Udp client dial with %v failed, err : %v\n", remoteAddr.String(), err.Error()))
	}
	defer conn.Close()

	if common.FlagInfos.ClientTimeoutSeconds != 0 {
		err = conn.SetDeadline(time.Now().Add(time.Duration(common.FlagInfos.ClientTimeoutSeconds) * time.Second))
		if err != nil {
			panic(err)
		}
	}

	common.Info("Udp client bind on %v, will sent data to %v\n", localAddr.String(), remoteAddr.String())

	if common.FlagInfos.ClientWaitingSeconds != 0 {
		common.Info("Udp client waiting %v...\n", common.FlagInfos.ClientWaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.ClientWaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Udp client[%v]----Udp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			continue
		}

		// receive
		recvBuffer := make([]byte, common.FlagInfos.ClientRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			common.Warn("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			continue
		}

		common.Info("Udp client[%v]----Udp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.FlagInfos.ClientSendData, recvBuffer[:n])
	}
}
