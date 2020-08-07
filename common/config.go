package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	CurrDir     string
	JsonConfigs *JsonConfigStruct
	FlagInfos   FlagInfoStruct
)

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
}

type JsonConfigStruct struct {
	// Common Config
	CommonLogLevel            int    `json:"CommonLogLevel"`
	CommonLogRoll             int    `json:"CommonLogRoll"`
	CommonRecvBufferSizeBytes uint64 `json:"CommonRecvBufferSizeBytes"`

	// Server Config
	ServerCounterOutputIntervalSeconds uint64   `json:ServerCounterOutputIntervalSeconds`
	ServerListenHost                   string   `json:"ServerListenHost"`
	ServerUdpListenHosts               []string `json:"ServerUdpListenHosts"`
	ServerTcpListenPorts               []uint16 `json:"ServerTcpListenPorts"`
	ServerUdpListenPorts               []uint16 `json:"ServerUdpListenPorts"`
	ServerHttpListenPorts              []uint16 `json:"ServerHttpListenPorts"`
	ServerHttpsListenPorts             []uint16 `json:"ServerHttpsListenPorts"`
	ServerQuicListenPorts              []uint16 `json:"ServerQuicListenPorts"`
	ServerDnsListenPorts               []uint16 `json:"ServerDnsListenPorts"`
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

type FlagInfoStruct struct {
	// common option
	IsHelp             bool
	IsServer           bool
	ConfigFileFullPath string
	// client option
	UsingTcp              bool
	UsingUdp              bool
	UsingHttp             bool
	UsingHttps            bool
	UsingQuic             bool
	UsingDns              bool
	ClientBindIpAddress   string
	ClientSendToIpAddress string
	SentToServerPort      uint16
	WaitingSeconds        uint64
	ClientSendNumbers     uint64
}

type IpAndPort struct {
	Ip   string
	Port uint16
}

func (addr *IpAndPort) String() string {
	if false == strings.Contains(FlagInfos.ClientBindIpAddress, ":") {
		return fmt.Sprintf("%v:%v", addr.Ip, addr.Port)
	} else {
		return fmt.Sprintf("[%v]:%v", addr.Ip, addr.Port)
	}
}

// 读取json配置文件
func LoadConfigFile(filePath string) (*JsonConfigStruct, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &JsonConfigStruct{}
	if err := json.Unmarshal(cData, c); nil != err {
		return nil, err
	}

	return c, nil
}
