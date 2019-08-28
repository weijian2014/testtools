package main

import (
	"../common"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"math/big"
)

func startIeeeQuicServer(listenPort uint16) {
	serverAddress := fmt.Sprintf("%v:%v", common.Configs.ServerListenHost, listenPort)
	listener, err := quic.ListenAddr(serverAddress, generateQuicTLSConfig(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Quic  server startup, listen on %v\n", serverAddress)

	for {
		session, err := listener.Accept()
		if err != nil {
			fmt.Printf("Quic server accept fail, err: %v\n", err)
			continue
		}

		go newQuicSessionHandler(session)
	}
}

func newQuicSessionHandler(sess quic.Session) {
	stream, err := sess.AcceptStream()
	defer stream.Close()
	if err != nil {
		fmt.Printf("Quic server[%v] ---- %v accept stream failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
		return
	}

	for {
		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		_, err = stream.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() {
				break
			}

			if "EOF" == err.Error() {
				break
			}

			fmt.Printf("Quic server[%v]----Quic client[%v] receive failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		// send
		n, err := stream.Write([]byte(common.Configs.ServerSendData))
		if nil != err {
			fmt.Printf("Quic server[%v]----Quic client[%v] send failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		fmt.Printf("Quic server[%v]----Quic client[%v]:\n\trecv: %s\n\tsend: %s\n", sess.LocalAddr(), sess.RemoteAddr(), recvBuffer[:n], common.Configs.ServerSendData)
	}

	fmt.Printf("Quic server[%v]----Quic client[%v] closed\n", sess.LocalAddr(), sess.RemoteAddr())
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
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}, NextProtos: []string{"ieee-quic"}}
}
