package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	CurrDir     string
	JsonConfigs *JsonConfig
	FlagInfos   FlagInfo
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

type JsonConfig struct {
	// Common Config
	CommonRecvBufferSizeBytes uint64 `json:"CommonRecvBufferSizeBytes"`
	// Server Config
	ServerListenHost            string `json:"ServerListenHost"`
	ServerTcpListenPort1        uint16 `json:"ServerTcpListenPort1"`
	ServerUdpListenPort1        uint16 `json:"ServerUdpListenPort1"`
	ServerHttpListenPort1       uint16 `json:"ServerHttpListenPort1"`
	ServerHttpsListenPort1      uint16 `json:"ServerHttpsListenPort1"`
	ServerGoogleQuicListenPort1 uint16 `json:"ServerGoogleQuicListenPort1"`
	ServerIeeeQuicListenPort1   uint16 `json:"ServerIeeeQuicListenPort1"`
	ServerTcpListenPort2        uint16 `json:"ServerTcpListenPort2"`
	ServerUdpListenPort2        uint16 `json:"ServerUdpListenPort2"`
	ServerHttpListenPort2       uint16 `json:"ServerHttpListenPort2"`
	ServerHttpsListenPort2      uint16 `json:"ServerHttpsListenPort2"`
	ServerGoogleQuicListenPort2 uint16 `json:"ServerGoogleQuicListenPort2"`
	ServerIeeeQuicListenPort2   uint16 `json:"ServerIeeeQuicListenPort2"`
	ServerDnsListenPort         uint16 `json:"ServerDnsListenPort"`
	// map[string]interface{}
	ServerDnsAEntrys  interface{} `json:"ServerDnsAEntrys"`
	ServerDns4AEntrys interface{} `json:"ServerDns4AEntrys"`
	ServerSendData    string      `json:"ServerSendData"`
	// Client Config
	ClientBindIpAddress          string `json:"ClientBindIpAddress"`
	ClientBindIpAddressRange     string `json:"ClientBindIpAddressRange"`
	ClientRangeModeChannelNumber uint64 `json:"ClientRangeModeChannelNumber"`
	ClientSendToIpv4Address      string `json:"ClientSendToIpv4Address"`
	ClientSendToIpv6Address      string `json:"ClientSendToIpv6Address"`
	ClientSendData               string `json:"ClientSendData"`
}

type FlagInfo struct {
	// common option
	IsHelp             bool
	IsServer           bool
	ConfigFileFullPath string
	// client option
	UsingTcp                      bool
	UsingUdp                      bool
	UsingHttp                     bool
	UsingHttps                    bool
	UsingGoogleQuic               bool
	UsingIEEEQuic                 bool
	UsingDns                      bool
	UsingClientBindIpAddressRange bool
	ClientBindIpAddress           string
	SentToServerPort              uint16
	WaitingSeconds                uint64
	ClientSendNumbers             uint64
}

// 读取json配置文件
func LoadConfigFile(filePath string) (*JsonConfig, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &JsonConfig{}
	if err := json.Unmarshal(cData, c); nil != err {
		return nil, err
	}

	return c, nil
}
