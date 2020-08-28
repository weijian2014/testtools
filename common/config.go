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
	ServerTcpListenHosts               []string `json:"ServerTcpListenHosts"`
	ServerUdpListenHosts               []string `json:"ServerUdpListenHosts"`
	ServerQuicListenHosts              []string `json:"ServerQuicListenHosts"`
	ServerHttpListenHosts              []string `json:"ServerHttpListenHosts"`
	ServerHttpsListenHosts             []string `json:"ServerHttpsListenHosts"`
	ServerDnsListenHosts               []string `json:"ServerDnsListenHosts"`
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
	str := ""
	if false == strings.Contains(addr.Ip, ":") {
		str = fmt.Sprintf("%v:%v", addr.Ip, addr.Port)
	} else {
		str = fmt.Sprintf("[%v]:%v", addr.Ip, addr.Port)
	}

	return str
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
