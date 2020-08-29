package server

import (
	"net"
	"runtime"
	"strings"
	"testtools/common"
)

func initTcpServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func() {
		common.Debug("%v server control coroutine running...\n", serverName)
		listener, err := net.Listen("tcp", listenAddr.String())
		if err != nil {
			panic(err)
		}

		c := make(chan int)
		err = insertControlChannel(listenAddr.String(), c)
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
					go tcpServerLoop(serverName, listener)
					isExit = false
				}
			case StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					listener.Close()
					err = deleteControlChannel(listenAddr.String())
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

func tcpServerLoop(serverName string, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				runtime.Goexit()
			} else {
				common.Warn("%v server accept failed, err: %v\n", serverName, err)
				continue
			}
		}

		go newTcpConnectionHandler(conn, serverName)
	}
}

func newTcpConnectionHandler(conn net.Conn, serverName string) {
	defer conn.Close()
	for {
		// receive
		recvBuffer := make([]byte, common.JsonConfigs.ServerRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() {
				break
			}

			if "EOF" == err.Error() {
				break
			}

			common.Warn("%v server[%v]----Tcp client[%v] receive failed, err: %v\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		// send
		_, err = conn.Write([]byte(common.JsonConfigs.ServerSendData))
		if nil != err {
			common.Warn("%v server[%v]----Tcp client[%v] send failed, err: %v\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		serverTcpCount++
		common.Info("%v server[%v]----Tcp client[%v]:\n\trecv: %s\n\tsend: %s\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), recvBuffer[:n], common.JsonConfigs.ServerSendData)
	}

	common.Info("%v server[%v]----Tcp client[%v] closed\n", serverName, conn.LocalAddr(), conn.RemoteAddr())
}
