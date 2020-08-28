package server

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testtools/common"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	StartServerControlOption int = iota // StartServer=0
	StopServerControlOption
	MaxControlOption
)

var (
	controlChannelsMap      map[uint16]chan int
	controlChannelsMapGuard sync.Mutex
)

func init() {
	controlChannelsMap = make(map[uint16]chan int)
}

func startConfigFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				{
					if !ok {
						common.Error("The %v file no ok in evnet!", common.FlagInfos.ConfigFileFullPath)
						return
					}

					if event.Name != common.FlagInfos.ConfigFileFullPath {
						break
					}

					if event.Op&fsnotify.Write == fsnotify.Write {
						common.Debug("The %v file has changed!\n", event.Name)
						err := reflushServers()
						if nil != err {
							common.Warn("The %v file watcher reflush servers fail for write, err: %v\n", common.FlagInfos.ConfigFileFullPath, err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				{
					if !ok {
						common.Error("The %v file no ok in error!", common.FlagInfos.ConfigFileFullPath)
						return
					}

					if nil != err {
						common.Error("The %v file watcher get error, err: %v\n", common.FlagInfos.ConfigFileFullPath, err)
						return
					}
				}
			}
		}
	}()

	common.System("The %v file watcher start ok", common.FlagInfos.ConfigFileFullPath)

	// watch the directory of file, but not a file
	err = watcher.Add(filepath.Dir(common.FlagInfos.ConfigFileFullPath))
	if err != nil {
		panic(err)
	}

	<-done
}

func reflushServers() error {
	newConfig, err := common.LoadConfigFile(common.FlagInfos.ConfigFileFullPath)
	if nil != err {
		return err
	}

	if reflect.DeepEqual(newConfig, common.FlagInfos) {
		return nil
	}

	// ServerLogLevel
	if common.JsonConfigs.ServerLogLevel != newConfig.ServerLogLevel {
		common.System("Server log level(%v) has changed, old=[%v-%v], new=[%v-%v]\n",
			common.GetLogLevel(), common.JsonConfigs.ServerLogLevel, common.LogLevelToString(common.JsonConfigs.ServerLogLevel),
			newConfig.ServerLogLevel, common.LogLevelToString(newConfig.ServerLogLevel))
		common.JsonConfigs.ServerLogLevel = newConfig.ServerLogLevel
		common.SetLogLevel(common.JsonConfigs.ServerLogLevel)
	}

	// ServerLogRoll
	if common.JsonConfigs.ServerLogRoll != newConfig.ServerLogRoll {
		common.System("Server log roll(line) has changed, old=[%v], new=[%v]\n", common.JsonConfigs.ServerLogRoll, newConfig.ServerLogRoll)
		common.JsonConfigs.ServerLogRoll = newConfig.ServerLogRoll
		common.SetLogRoll(common.JsonConfigs.ServerLogRoll)
	}

	// ServerRecvBufferSizeBytes
	if common.JsonConfigs.ServerRecvBufferSizeBytes != newConfig.ServerRecvBufferSizeBytes {
		common.System("Server recvice buffer size(byte) has changed, old=[%v], new=[%v]\n", common.JsonConfigs.ServerRecvBufferSizeBytes, newConfig.ServerRecvBufferSizeBytes)
		common.JsonConfigs.ServerRecvBufferSizeBytes = newConfig.ServerRecvBufferSizeBytes
	}

	// ServerCounterOutputIntervalSeconds
	if common.JsonConfigs.ServerCounterOutputIntervalSeconds != newConfig.ServerCounterOutputIntervalSeconds {
		common.System("Server counter output interval(second) seconds has changed, old=[%v], new=[%v]\n", common.JsonConfigs.ServerCounterOutputIntervalSeconds, newConfig.ServerCounterOutputIntervalSeconds)
		common.JsonConfigs.ServerCounterOutputIntervalSeconds = newConfig.ServerCounterOutputIntervalSeconds
	}

	// ServerSendData
	if common.JsonConfigs.ServerSendData != newConfig.ServerSendData {
		common.System("Server send data has changed, old=[%v], new=[%v]\n", common.JsonConfigs.ServerSendData, newConfig.ServerSendData)
		common.JsonConfigs.ServerSendData = newConfig.ServerSendData
	}

	// reflush tcp server
	err = reflushServer(common.JsonConfigs.ServerTcpListenHosts, newConfig.ServerTcpListenHosts, "Tcp")
	if nil != err {
		return err
	} else {
		common.JsonConfigs.ServerTcpListenHosts = newConfig.ServerTcpListenHosts
	}

	// reflush udp server
	err = reflushServer(common.JsonConfigs.ServerUdpListenHosts, newConfig.ServerUdpListenHosts, "Udp")
	if nil != err {
		return err
	} else {
		common.JsonConfigs.ServerUdpListenHosts = newConfig.ServerUdpListenHosts
	}

	// reflush quic server
	err = reflushServer(common.JsonConfigs.ServerQuicListenHosts, newConfig.ServerQuicListenHosts, "Quic")
	if nil != err {
		return err
	} else {
		common.JsonConfigs.ServerQuicListenHosts = newConfig.ServerQuicListenHosts
	}

	// reflush http server
	err = reflushServer(common.JsonConfigs.ServerHttpListenHosts, newConfig.ServerHttpListenHosts, "Http")
	if nil != err {
		return err
	} else {
		common.JsonConfigs.ServerHttpListenHosts = newConfig.ServerHttpListenHosts
	}

	// reflush http server
	err = reflushServer(common.JsonConfigs.ServerHttpsListenHosts, newConfig.ServerHttpsListenHosts, "Https")
	if nil != err {
		return err
	} else {
		common.JsonConfigs.ServerHttpsListenHosts = newConfig.ServerHttpsListenHosts
	}

	// reflush dns entry
	if !reflect.DeepEqual(newConfig.ServerDnsAEntrys, common.JsonConfigs.ServerDnsAEntrys) ||
		!reflect.DeepEqual(newConfig.ServerDns4AEntrys, common.JsonConfigs.ServerDns4AEntrys) {
		common.System("Dns server entrys has changed, oldA=%v, old4A=%v, newA=%v, new4A=%v\n",
			common.JsonConfigs.ServerDnsAEntrys, common.JsonConfigs.ServerDns4AEntrys, newConfig.ServerDnsAEntrys, newConfig.ServerDns4AEntrys)
		common.JsonConfigs.ServerDnsAEntrys = newConfig.ServerDnsAEntrys
		common.JsonConfigs.ServerDns4AEntrys = newConfig.ServerDns4AEntrys
		saveDnsEntrys()
		printDnsServerEntrys()
		common.System("\n")
	}

	common.JsonConfigs = newConfig
	return nil
}

func insertControlChannel(port uint16, ctrlChan chan int) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	_, ok := controlChannelsMap[port]
	if ok {
		return fmt.Errorf("The control channel already exists")
	}

	controlChannelsMap[port] = ctrlChan
	return nil
}

func deleteControlChannel(port uint16) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	_, ok := controlChannelsMap[port]
	if !ok {
		return fmt.Errorf("The control channel does not exist")
	}

	delete(controlChannelsMap, port)
	return nil
}

func sendOptionToControlChannel(port uint16, option int) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	err := checkControlOption(option)
	if nil != err {
		return err
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	ctrlChan, ok := controlChannelsMap[port]
	if !ok {
		return fmt.Errorf("The control channel does not exist")
	}

	ctrlChan <- option
	return nil
}

func startAllServers() error {
	for port, _ := range controlChannelsMap {
		err := sendOptionToControlChannel(port, StartServerControlOption)
		if nil != err {
			return err
		}

		time.Sleep(time.Duration(50) * time.Millisecond)
	}

	return nil
}

func stopAllServers() error {
	for port, _ := range controlChannelsMap {
		err := sendOptionToControlChannel(port, StopServerControlOption)
		if nil != err {
			return err
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return nil
}

func startServer(port uint16) error {
	return sendOptionToControlChannel(port, StartServerControlOption)
}

func stopServer(port uint16) error {
	return sendOptionToControlChannel(port, StopServerControlOption)
}

func checkControlOption(option int) error {
	if StartServerControlOption > option || MaxControlOption <= option {
		return fmt.Errorf("Invalid option")
	}

	return nil
}

func intersection(aSet, bSet []string) []string {
	r := make([]string, 0)
	for _, a := range aSet {
		for _, b := range bSet {
			if a == b {
				r = append(r, a)
			}
		}
	}

	return r
}

func subtraction(aSet, bSet []string) []string {
	s := make([]string, 0)

	isSub := true
	for _, a := range aSet {
		isSub = true
		for _, b := range bSet {
			if a == b {
				isSub = false
				break
			}
		}

		if isSub {
			s = append(s, a)
		}
	}

	return s
}

func reflushServer(old, new []string, name string) error {
	if !reflect.DeepEqual(new, old) {
		r := intersection(new, old)
		del := subtraction(old, r)
		add := subtraction(new, r)

		if 0 != len(del) || 0 != len(add) {
			common.System("%v server listen hosts has changed, old=%v, new=%v, del=%v, add=%v\n", name, old, old, del, add)

			for _, host := range del {
				_, port, err := common.GetIpAndPort(host)
				if nil != err {
					return err
				}

				err = stopServer(port)
				if nil != err {
					common.Error("The %v file watcher stop [%vServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, name, port, err)
					return err
				}
			}

			listenAddr := common.IpAndPort{Ip: "0.0.0.0", Port: 0}
			for _, host := range add {
				ip, port, err := common.GetIpAndPort(host)
				if nil != err {
					return err
				}

				listenAddr.Ip = ip
				listenAddr.Port = port
				initTcpServer(fmt.Sprintf("%vServer-%v", name, port), listenAddr)
				time.Sleep(time.Duration(500) * time.Millisecond)
				err = startServer(port)
				if nil != err {
					common.Error("The %v file watcher start [%vServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, name, port, err)
					return err
				}
			}
		}
	}

	return nil
}
