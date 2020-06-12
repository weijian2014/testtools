package server

import (
	"net"
	"testtools/common"
)

func startTcpServer(serverName string, listenAddr *common.IpAndPort) {
	listener, err := net.Listen("tcp", listenAddr.String())
	if err != nil {
		panic(err)
	}

	common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())

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
