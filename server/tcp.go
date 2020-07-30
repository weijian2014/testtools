package server

import (
	"net"
	"runtime"
	"testtools/common"
)

func initTcpServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	f := func(serverName string, listenAddr common.IpAndPort) {
		common.Debug("%v server control coroutine running...\n", serverName)
		listener, err := net.Listen("tcp", listenAddr.String())
		if err != nil {
			panic(err)
		}

		c := make(chan int)
		err = common.InsertControlChannel(listenAddr.Port, c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case common.StartServerControlOption:
				{
					common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
					go tcpServerLoop(serverName, listener)
					isExit = false
					continue
				}
			case common.StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					listener.Close()
					err = common.DeleteControlChannel(listenAddr.Port)
					if nil != err {
						common.Error("Delete control channel fial, erro: %v", err)
					}
					isExit = true
					break
				}
			default:
				{
					isExit = false
					continue
				}
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}

	go f(serverName, listenAddr)
}

func tcpServerLoop(serverName string, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			common.Warn("%v server accept failed, err: %v\n", serverName, err)
			continue
		}

		go newTcpConnectionHandler(conn, serverName)
	}
}

func newTcpConnectionHandler(conn net.Conn, serverName string) {
	defer conn.Close()
	for {
		// receive
		recvBuffer := make([]byte, common.JsonConfigs.CommonRecvBufferSizeBytes)
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

		common.Info("%v server[%v]----Tcp client[%v]:\n\trecv: %s\n\tsend: %s\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), recvBuffer[:n], common.JsonConfigs.ServerSendData)
	}

	serverTcpCount++
	common.Info("%v server[%v]----Tcp client[%v] closed\n", serverName, conn.LocalAddr(), conn.RemoteAddr())
}
