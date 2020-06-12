package client

import (
	"fmt"
	"net"
	"testtools/common"
	"time"
)

func sendByTcp(localAddr, remoteAddr *common.IpAndPort) {
	lAddr, err := net.ResolveTCPAddr("tcp", localAddr.String())
	if nil != err {
		panic(err)
	}

	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr.String())
	if nil != err {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", lAddr, rAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Tcp client connect to %v failed, err : %v\n", remoteAddr.String(), err.Error()))
	}

	common.Info("Tcp client bind on %v, will sent data to %v\n", localAddr.String(), remoteAddr.String())

	if 0 != common.FlagInfos.WaitingSeconds {
		common.Info("Tcp client waiting %v...\n", common.FlagInfos.WaitingSeconds)
		time.Sleep(time.Duration(common.FlagInfos.WaitingSeconds) * time.Second)
	}
	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(common.JsonConfigs.ClientSendData))
		if err != nil {
			common.Warn("Tcp client[%v]----Tcp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			continue
		}

		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			common.Warn("Tcp client[%v]----Tcp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			continue
		}

		common.Info("Tcp client[%v]----Tcp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, common.JsonConfigs.ClientSendData, recvBuffer[:n])
	}
}
