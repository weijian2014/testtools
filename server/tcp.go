package main

import (
	"fmt"
	"net"
	"testtools/common"
)

func startTcpServer(listenPort uint16) {
	serverAddress := fmt.Sprintf("%v:%v", common.Configs.ServerListenHost, listenPort)
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Tcp   server startup, listen on %v\n", serverAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Tpc server accept failed, err: %v\n", err)
			continue
		}

		go newTcpConnectionHandler(conn)
	}
}

func newTcpConnectionHandler(conn net.Conn) {
	defer conn.Close()
	for {
		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() {
				break
			}

			if "EOF" == err.Error() {
				break
			}

			fmt.Printf("Tcp server[%v]----Tcp client[%v] receive failed, err: %v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		// send
		_, err = conn.Write([]byte(common.Configs.ServerSendData))
		if nil != err {
			fmt.Printf("Tcp server[%v]----Tcp client[%v] send failed, err: %v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		fmt.Printf("Tcp server[%v]----Tcp client[%v]:\n\trecv: %s\n\tsend: %s\n", conn.LocalAddr(), conn.RemoteAddr(), recvBuffer[:n], common.Configs.ServerSendData)
	}

	fmt.Printf("Tcp server[%v]----Tcp client[%v] closed\n", conn.LocalAddr(), conn.RemoteAddr())
}
