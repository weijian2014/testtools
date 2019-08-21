package main

import (
	"../common"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	"net"
	"syscall"
	"time"
)

func sendIosByTcp() {
	defer util.Run()()

	localAddr := net.ParseIP(common.Configs.ClientBindIpAddress)
	if nil == localAddr {
		panic(errors.New("local address error"))
	}
	localAddr = localAddr.To4()

	remoteAddr := net.ParseIP(sendToServerIpAddress)
	if nil == remoteAddr {
		panic(errors.New("remote address error"))
	}
	remoteAddr = remoteAddr.To4()

	ipLayer := layers.IPv4{
		SrcIP:    localAddr,
		DstIP:    remoteAddr,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	tcpOptions := []layers.TCPOption{
		{
			OptionType:   layers.TCPOptionKindMSS,
			OptionLength: 4,
			OptionData:   []byte("1460"),
		},
		{
			OptionType: layers.TCPOptionKindNop,
		},
		{
			OptionType:   layers.TCPOptionKindWindowScale,
			OptionLength: 3,
			OptionData:   []byte("128"),
		},
		{
			OptionType: layers.TCPOptionKindNop,
		},
		{
			OptionType: layers.TCPOptionKindNop,
		},
		{
			OptionType:   layers.TCPOptionKindTimestamps,
			OptionLength: 10,
			OptionData:   []byte("2320491952"),
		},
		{
			OptionType:   layers.TCPOptionKindSACKPermitted,
			OptionLength: 2,
		},
		{
			OptionType: layers.TCPOptionKindEndList,
		},
	}

	tcpLayer := layers.TCP{
		SrcPort: layers.TCPPort(5566),
		DstPort: layers.TCPPort(common.Configs.ServerTcpListenPort),
		SYN:     true,
		Window:  65535,
		Urgent:  0,
		Options: tcpOptions,
	}

	payload := gopacket.Payload([]byte("meowmeowmeowXXXhoho"))
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		//FixLengths:       true,
		ComputeChecksums: true,
	}

	tcpLayer.SetNetworkLayerForChecksum(&ipLayer)
	err := gopacket.SerializeLayers(buf, opts, &ipLayer, &tcpLayer, payload)
	if err != nil {
		panic(err)
	}

	//ipConn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	//if err != nil {
	//	panic(err)
	//}
	//
	//_, err = ipConn.WriteTo(buf.Bytes(), &net.IPAddr{IP: remoteAddr})
	//if err != nil {
	//	panic(err)
	//}

	sockfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		panic(err)
	}
	defer syscall.Shutdown(sockfd, syscall.SHUT_RDWR)

	var remoteIpPort syscall.SockaddrInet4
	remoteIpPort.Addr = [4]byte{127, 0, 0, 1}
	remoteIpPort.Port = int(common.Configs.ServerTcpListenPort)
	syscall.Sendto(sockfd, buf.Bytes(), 0, &remoteIpPort)

	fmt.Printf("IP layer:%+v\n", gopacket.LayerString(&ipLayer))
	fmt.Printf("IP layer:%+v\n", gopacket.LayerString(&tcpLayer))
	time.Sleep(time.Duration(2) * time.Second)
}
