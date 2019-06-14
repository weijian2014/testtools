package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/net/dns/dnsmessage"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	CurrDir               = ""
	IsHelp                = false
	UsingTcp              bool
	UsingUdp              bool
	UsingHttp             bool
	UsingHttps            bool
	UsingDns              bool
	UsingIEEEQuic         bool
	Configs               *Config = nil
	ConfigFileFullPath            = ""
	ClientBindIpAddress           = ""
	ClientSendNumbers             = 1
	SendToServerIpAddress string
)

type Config struct {
	// Common Config
	CommonRecvBufferSizeBytes uint64 `json:"CommonRecvBufferSizeBytes"`
	// Server Config
	ServerListenHost           string `json:"ServerListenHost"`
	ServerTcpListenPort        uint16 `json:"ServerTcpListenPort"`
	ServerUdpListenPort        uint16 `json:"ServerUdpListenPort"`
	ServerHttpListenPort       uint16 `json:"ServerHttpListenPort"`
	ServerHttpsListenPort      uint16 `json:"ServerHttpsListenPort"`
	ServerIeeeQuicListenPort   uint16 `json:"ServerIeeeQuicListenPort"`
	ServerGoogleQuicListenPort uint16 `json:"ServerGoogleQuicListenPort"`
	ServerDnsListenPort        uint16 `json:"ServerDnsListenPort"`
	// map[string]interface{}
	ServerDnsAEntrys  interface{} `json:"ServerDnsAEntrys"`
	ServerDns4AEntrys interface{} `json:"ServerDns4AEntrys"`
	ServerSendData    string      `json:"ServerSendData"`
	// Client Config
	ClientBindIpAddress     string `json:"ClientBindIpAddress"`
	ClientSendToIpv4Address string `json:"ClientSendToIpv4Address"`
	ClientSendToIpv6Address string `json:"ClientSendToIpv6Address"`
	ClientSendData          string `json:"ClientSendData"`
}

func init() {
	currFullPath, err := exec.LookPath(os.Args[0])
	if nil != err {
		panic(err)
	}

	absFullPath, err := filepath.Abs(currFullPath)
	if nil != err {
		panic(err)
	}
	CurrDir = filepath.Dir(absFullPath)

	flag.BoolVar(&IsHelp, "h", false, "Show help")
	flag.StringVar(&ClientBindIpAddress, "b", "", "The ip address of client bind\n"+
		"This parameter takes precedence over ClientBindIpAddress in the config.json file\n"+
		"If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file")
	flag.IntVar(&ClientSendNumbers, "n", 1, "The number of client send data to server, valid only for UDP, TCP, iQuic and gQuic protocols")
	flag.StringVar(&ConfigFileFullPath, "f", CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.BoolVar(&UsingTcp, "tcp", true, "Using TCP protocol")
	flag.BoolVar(&UsingUdp, "udp", false, "Using UDP protocol")
	flag.BoolVar(&UsingHttp, "http", false, "Using HTTP protocol")
	flag.BoolVar(&UsingHttps, "https", false, "Using HTTPS protocol")
	flag.BoolVar(&UsingIEEEQuic, "iquic", false, "Using IEEE QUIC protocol")
	flag.BoolVar(&UsingDns, "dns", false, "Using DNS protocol")
	flag.Parse()

	if UsingUdp || UsingHttps || UsingHttp || UsingDns || UsingIEEEQuic {
		UsingTcp = false
	}
}

func main() {
	var err error
	_, err = os.Stat(ConfigFileFullPath)
	if os.IsNotExist(err) {
		ConfigFileFullPath = CurrDir + "/../config/config.json"
	}
	Configs, err = loadConfigFile(ConfigFileFullPath)
	if nil != err {
		panic(err)
	}
	err = checkConfigFlie()
	if nil != err {
		panic(err)
	}

	if nil != net.ParseIP(ClientBindIpAddress) {
		Configs.ClientBindIpAddress = ClientBindIpAddress
	}

	if IsHelp {
		flag.Usage()
		fmt.Printf("\nJson config: %+v\n\n", Configs)
		return
	}

	if false == strings.Contains(Configs.ClientBindIpAddress, ":") {
		SendToServerIpAddress = Configs.ClientSendToIpv4Address
	} else {
		SendToServerIpAddress = Configs.ClientSendToIpv6Address
	}

	if UsingTcp {
		sendByTcp()
		return
	}

	if UsingUdp {
		sendByUdp()
		return
	}

	if UsingHttp {
		sendByHttp()
		return
	}

	if UsingHttps {
		sendByHttps()
		return
	}

	if UsingDns {
		sendByDns()
		return
	}

	if UsingIEEEQuic {
		sendByIEEEQuic()
		return
	}
}

func sendByTcp() {
	localAddr := &net.TCPAddr{IP: net.ParseIP(Configs.ClientBindIpAddress)}
	remoteAddr := &net.TCPAddr{IP: net.ParseIP(SendToServerIpAddress), Port: int(Configs.ServerTcpListenPort)}
	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Tcp client connect to %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Tcp client bind on %v, will sent data to %v\n", Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Tcp client[%v]----Udp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Tcp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Tcp client[%v]----Tcp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, Configs.ClientSendData, recvBuffer[:n])
	}
}

func sendByUdp() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(Configs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(SendToServerIpAddress), Port: int(Configs.ServerUdpListenPort)}
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Udp client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	fmt.Printf("Udp client bind on %v, will sent data to %v\n", Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send
		_, err = conn.Write([]byte(Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("Udp client[%v]----Udp server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", conn.LocalAddr(), conn.RemoteAddr(), i, Configs.ClientSendData, recvBuffer[:n])
	}
}

func sendByHttp() {
	localAddr, err := net.ResolveIPAddr("ip", Configs.ClientBindIpAddress)
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var reqeustUrl string
	if true == strings.Contains(SendToServerIpAddress, ".") {
		reqeustUrl = fmt.Sprintf("http://%s:%d", SendToServerIpAddress, Configs.ServerHttpListenPort)
	} else {
		reqeustUrl = fmt.Sprintf("http://[%s]:%d", SendToServerIpAddress, Configs.ServerHttpListenPort)
	}

	fmt.Printf("Http client bind on %v, will reqeust to %v\n", Configs.ClientBindIpAddress, reqeustUrl)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Http client new request failed, times[%d], err : %v\n", i, err.Error())
			return
		}
		req.Header.Add("ClientSendData", Configs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Http client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			fmt.Printf("Http client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		fmt.Printf("Http client[%v]----Http server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, Configs.ClientSendData, body)
	}
}

func sendByHttps() {
	localAddr, err := net.ResolveIPAddr("ip", Configs.ClientBindIpAddress)
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略证书
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var reqeustUrl string
	if true == strings.Contains(SendToServerIpAddress, ".") {
		reqeustUrl = fmt.Sprintf("https://%s:%d", SendToServerIpAddress, Configs.ServerHttpsListenPort)
	} else {
		reqeustUrl = fmt.Sprintf("https://[%s]:%d", SendToServerIpAddress, Configs.ServerHttpsListenPort)
	}

	fmt.Printf("Https client bind on %v, will reqeust to %v\n", Configs.ClientBindIpAddress, reqeustUrl)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Https client new request failed, times[%d], err : %v\n", i, err.Error())
			return
		}
		req.Header.Add("ClientSendData", Configs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Https client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			fmt.Printf("Https client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		fmt.Printf("Https client[%v]----Https server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, Configs.ClientSendData, body)
	}
}

func sendByDns() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(Configs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(SendToServerIpAddress), Port: int(Configs.ServerDnsListenPort)}
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	defer conn.Close()
	if err != nil {
		panic(fmt.Sprintf("Dns client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	var questionType dnsmessage.Type
	if false == strings.Contains(Configs.ClientBindIpAddress, ":") {
		questionType = dnsmessage.TypeA
	} else {
		questionType = dnsmessage.TypeAAAA
	}
	requestMessage := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 8888,
			Response:           false,
			OpCode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: false,
			RCode:              dnsmessage.RCodeSuccess,
		},
		Questions: []dnsmessage.Question{
			{
				Name:  mustNewName("www.example.com."),
				Type:  questionType,
				Class: dnsmessage.ClassINET,
			},
		},
	}

	fmt.Printf("Dns client bind on %v, will sent query to %v\n", Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			fmt.Printf("Dns client[%v]----Dns server[%v] pack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}
		_, err = conn.Write(packed)
		if err != nil {
			fmt.Printf("Dns client[%v]----Dns server[%v] send failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}

		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		_, err = conn.Read(recvBuffer)
		if err != nil {
			fmt.Printf("Udp client[%v]----Udp server[%v] receive failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err.Error())
			return
		}
		var responseMessage dnsmessage.Message
		err = responseMessage.Unpack(recvBuffer)
		if nil != err {
			fmt.Printf("Dns client[%v]----Dns server[%v] unpack failed, times[%d], err : %v\n", conn.LocalAddr(), conn.RemoteAddr(), i, err)
			return
		}

		if dnsmessage.TypeA == questionType {
			ipv4 := responseMessage.Answers[0].Body.GoString()
			fmt.Printf("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\tanswers: %+v\n",
				conn.LocalAddr(), conn.RemoteAddr(), i, requestMessage.Questions[0], ipv4)
		} else {
			ipv6 := responseMessage.Answers[0].Body.GoString()
			fmt.Printf("Dns client[%v]----Dns server[%v], times[%d]:\n\tquestion: %+v\n\t answers: %+v\n",
				conn.LocalAddr(), conn.RemoteAddr(), i, requestMessage.Questions[0], ipv6)
		}
	}
}

func sendByIEEEQuic() {
	localAddr := &net.UDPAddr{IP: net.ParseIP(Configs.ClientBindIpAddress)}
	remoteAddr := &net.UDPAddr{IP: net.ParseIP(SendToServerIpAddress), Port: int(Configs.ServerIeeeQuicListenPort)}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: localAddr.IP, Port: localAddr.Port})
	if err != nil {
		panic(err)
	}

	session, err := quic.Dial(udpConn, remoteAddr, remoteAddr.String(), &tls.Config{InsecureSkipVerify: true}, nil)
	defer session.Close()
	if err != nil {
		panic(fmt.Sprintf("IEEE Quic client dial with %v failed, err : %v\n", remoteAddr, err.Error()))
	}

	stream, err := session.OpenStreamSync()
	defer stream.Close()
	if err != nil {
		panic(fmt.Sprintf("IEEE Quic client[%v]----Quic server[%v] open stream failed, err : %v\n", session.LocalAddr(), session.RemoteAddr(), err.Error()))
		return
	}

	fmt.Printf("IEEE Quic client bind on %v, will sent data to %v\n", Configs.ClientBindIpAddress, remoteAddr)
	for i := 1; i <= ClientSendNumbers; i++ {
		// send
		_, err = stream.Write([]byte(Configs.ClientSendData))
		if err != nil {
			fmt.Printf("IEEE Quic client[%v]----Quic server[%v] send failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		n, err := stream.Read(recvBuffer)
		if err != nil {
			fmt.Printf("IEEE Quic client[%v]----Quic server[%v] receive failed, times[%d], err : %v\n", session.LocalAddr(), session.RemoteAddr(), i, err.Error())
			return
		}

		fmt.Printf("IEEE Quic client[%v]----Quic server[%v], times[%d]:\n\tsend: %s\n\trecv: %s\n", session.LocalAddr(), session.RemoteAddr(), i, Configs.ClientSendData, recvBuffer[:n])
	}
}

// 读取json配置文件
func loadConfigFile(filePath string) (*Config, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := json.Unmarshal(cData, c); nil != err {
		return nil, err
	}

	return c, nil
}

func checkConfigFlie() error {
	if 0 != len(ClientBindIpAddress) &&
		nil == net.ParseIP(ClientBindIpAddress) {
		return errors.New(fmt.Sprintf("ClientBindIpAddress[%v] is invalid address, please check -b option", ClientBindIpAddress))
	}

	if nil == net.ParseIP(Configs.ClientSendToIpv4Address) ||
		false == strings.Contains(Configs.ClientSendToIpv4Address, ".") {
		return errors.New(fmt.Sprintf("ClientSendToIpv4Address[%v] is invalid ipv4 address in the config.json file", Configs.ClientSendToIpv4Address))
	}

	if nil == net.ParseIP(Configs.ClientSendToIpv6Address) ||
		false == strings.Contains(Configs.ClientSendToIpv6Address, ":") {
		return errors.New(fmt.Sprintf("ClientSendToIpv6Address[%v] is invalid ipv6 address in the config.json file", Configs.ClientSendToIpv6Address))
	}

	return nil
}

func isLocalIP(ip string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for i := range addrs {
		intf, _, err := net.ParseCIDR(addrs[i].String())
		if err != nil {
			return false, err
		}
		if net.ParseIP(ip).Equal(intf) {
			return true, nil
		}
	}
	return false, nil
}

func mustNewName(name string) dnsmessage.Name {
	n, err := dnsmessage.NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}
