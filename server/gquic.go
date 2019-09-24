package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"math/big"
	"testtools/common"
)

func startGQuicServer(listenPort uint16, serverName string) {
	serverAddress := fmt.Sprintf("%v:%v", common.Configs.ServerListenHost, listenPort)
	listener, err := quic.ListenAddr(serverAddress, generateQuicTLSConfig(), &quic.Config{
		Versions: []quic.VersionNumber{
			quic.VersionGQUIC39,
			quic.VersionGQUIC43,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v server startup, listen on %v\n", serverName, serverAddress)

	for {
		session, err := listener.Accept()
		if err != nil {
			fmt.Printf("%v server accept fail, err: %v\n", serverName, err)
			continue
		}

		go newQuicSessionHandler(session, serverName)
	}
}

func newQuicSessionHandler(sess quic.Session, serverName string) {
	stream, err := sess.AcceptStream()
	defer stream.Close()
	if err != nil {
		fmt.Printf("%v server[%v] ---- %v accept stream failed, err: %v\n", serverName, sess.LocalAddr(), sess.RemoteAddr(), err)
		return
	}

	for {
		// receive
		recvBuffer := make([]byte, common.Configs.CommonRecvBufferSizeBytes)
		_, err = stream.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() || "EOF" == err.Error() || "PeerGoingAway:" == err.Error() {
				break
			}

			fmt.Printf("%v server[%v]----Quic client[%v] receive failed, err: %v\n", serverName, sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		// send
		n, err := stream.Write([]byte(common.Configs.ServerSendData))
		if nil != err {
			fmt.Printf("%v server[%v]----Quic client[%v] send failed, err: %v\n", serverName, sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		fmt.Printf("%v server[%v]----Quic client[%v]:\n\trecv: %s\n\tsend: %s\n",
			serverName, sess.LocalAddr(), sess.RemoteAddr(), recvBuffer[:n], common.Configs.ServerSendData)
	}

	fmt.Printf("%v server[%v]----Quic client[%v] closed\n", serverName, sess.LocalAddr(), sess.RemoteAddr())
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
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
