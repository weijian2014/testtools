package server

import (
	"net"
	"runtime"
	"strings"
	"testtools/common"
)

func initUdpServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func() {
		common.Debug("%v server control coroutine running...\n", serverName)
		lAddr, err := net.ResolveUDPAddr("udp", listenAddr.String())
		if err != nil {
			panic(err)
		}

		conn, err := net.ListenUDP("udp", lAddr)
		defer conn.Close()
		if err != nil {
			panic(err)
		}

		c := make(chan int)
		err = insertControlChannel(listenAddr.Port, c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case StartServerControlOption:
				{
					common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
					go udpServerLoop(serverName, conn)
					isExit = false
				}
			case StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					conn.Close()
					err = deleteControlChannel(listenAddr.Port)
					if nil != err {
						common.Error("Delete control channel fial, erro: %v", err)
					}
					isExit = true
				}
			default:
				{
					isExit = false
				}
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}()
}

func udpServerLoop(serverName string, conn *net.UDPConn) {
	for {
		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				runtime.Goexit()
			} else {
				common.Warn("%v server[%v]----Udp client[%v] receive failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
				continue
			}
		}

		// send
		n, err := conn.WriteToUDP([]byte(common.JsonConfigs.ServerSendData), remoteAddress)
		if err != nil {
			common.Warn("%v server[%v]----Udp client[%v] send failed, err : %v\n", serverName, conn.LocalAddr(), remoteAddress, err)
			continue
		}

		serverUdpCount++
		common.Info("%v server[%v]----Udp client[%v]:\n\trecv: %s\n\tsend: %s\n", serverName, conn.LocalAddr(), remoteAddress, recvBuffer[:n], common.JsonConfigs.ServerSendData)
	}
}
