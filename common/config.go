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
	ServerLogLevel                     int         `json:"ServerLogLevel"`
	ServerLogRoll                      int         `json:"ServerLogRoll"`
	ServerRecvBufferSizeBytes          uint64      `json:"ServerRecvBufferSizeBytes"`
	ServerCounterOutputIntervalSeconds uint64      `json:ServerCounterOutputIntervalSeconds`
	ServerTcpListenHosts               []string    `json:"ServerTcpListenHosts"`
	ServerUdpListenHosts               []string    `json:"ServerUdpListenHosts"`
	ServerQuicListenHosts              []string    `json:"ServerQuicListenHosts"`
	ServerHttpListenHosts              []string    `json:"ServerHttpListenHosts"`
	ServerHttpsListenHosts             []string    `json:"ServerHttpsListenHosts"`
	ServerHttp20ListenHosts            []string    `json:"ServerHttp20ListenHosts"`
	ServerDnsListenHosts               []string    `json:"ServerDnsListenHosts"`
	ServerDnsAEntrys                   interface{} `json:"ServerDnsAEntrys"`  // map[string]interface{}
	ServerDns4AEntrys                  interface{} `json:"ServerDns4AEntrys"` // map[string]interface{}
	ServerSendData                     string      `json:"ServerSendData"`
}

type FlagInfoStruct struct {
	IsHelp             bool
	IsServer           bool
	ConfigFileFullPath string
	// client option
	ClientUsingTcp            bool
	ClientUsingUdp            bool
	ClientUsingHttp           bool
	ClientUsingHttps          bool
	ClientUsingHttp20         bool
	ClientUsingQuic           bool
	ClientUsingDns            bool
	ClientScrIp               string
	ClientDestIp              string
	ClientDestPort            uint16
	ClientWaitingSeconds      uint64
	ClientSendNumbers         uint64
	ClientLogLevel            int
	ClientSendData            string
	ClientRecvBufferSizeBytes uint64
	ClientQuicAlpn            []string
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
