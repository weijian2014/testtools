package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"runtime"
	"strings"
	"testtools/common"

	"github.com/quic-go/quic-go"
)

func initQuicServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func() {
		common.Debug("%v server control coroutine running...\n", serverName)
		listener, err := quic.ListenAddr(listenAddr.String(), generateQuicTLSConfig(), nil)
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
					go quicServerLoop(serverName, listener)
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

func quicServerLoop(serverName string, listener *quic.Listener) {
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "server closed") {
				runtime.Goexit()
			} else {
				common.Warn("%v server accept fail, err: %v\n", serverName, err)
				continue
			}
		}

		go newQuicSessionHandler(conn, serverName)
	}
}

func newQuicSessionHandler(conn quic.Connection, serverName string) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		common.Warn("%v server[%v] ---- %v accept stream failed, err: %v\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), err)
		return
	}
	defer stream.Close()

	for {
		// receive
		recvBuffer := make([]byte, common.JsonConfigs.ServerRecvBufferSizeBytes)
		_, err = stream.Read(recvBuffer)
		if err != nil {
			if err.Error() == "NO_ERROR" ||
				err.Error() == "EOF" ||
				strings.Contains(err.Error(), "PeerGoingAway") ||
				strings.Contains(err.Error(), "NetworkIdleTimeout") ||
				strings.Contains(err.Error(), "No recent network activity") {
				break
			}

			common.Warn("%v server[%v]----Quic client[%v] receive failed, err: %v\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), err)
			return
		}

		// send
		n, err := stream.Write([]byte(common.JsonConfigs.ServerSendData))
		if nil != err {
			common.Warn("%v server[%v]----Quic client[%v] send failed, err: %v\n", serverName, conn.LocalAddr(), conn.RemoteAddr(), err)
			return
		}

		serverQuicCount++
		common.Info("%v server[%v]----Quic client[%v]:\n\trecv: %s\n\tsend: %s\n",
			serverName, conn.LocalAddr(), conn.RemoteAddr(), recvBuffer[:n], common.JsonConfigs.ServerSendData)
	}

	common.Info("%v server[%v]----Quic client[%v] closed\n", serverName, conn.LocalAddr(), conn.RemoteAddr())
}

// Setup a bare-bones TLS config for the server
func generateQuicTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{common.COMMON_QUIC_ALPN},
	}
}
