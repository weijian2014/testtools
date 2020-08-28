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
						common.Debug("The %v file has changed!\n", event.Name)
						break
					}

					if event.Op&fsnotify.Write == fsnotify.Write {
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
			common.GetLogLevel(),
			common.JsonConfigs.ServerLogLevel, common.LogLevelToString(common.JsonConfigs.ServerLogLevel),
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

	// listenAddr := common.IpAndPort{Ip: "0.0.0.0", Port: 0}

	// // reflush tcp server
	// if !reflect.DeepEqual(newConfig.ServerTcpListenPorts, common.JsonConfigs.ServerTcpListenPorts) {
	// 	r := intersection(newConfig.ServerTcpListenPorts, common.JsonConfigs.ServerTcpListenPorts)
	// 	del := subtraction(common.JsonConfigs.ServerTcpListenPorts, r)
	// 	add := subtraction(newConfig.ServerTcpListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Tcp server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerTcpListenPorts, newConfig.ServerTcpListenPorts, del, add)
	// 		common.JsonConfigs.ServerTcpListenPorts = newConfig.ServerTcpListenPorts

	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [TcpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			initTcpServer(fmt.Sprintf("TcpServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(200) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [TcpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// // reflush udp server
	// if !reflect.DeepEqual(newConfig.ServerUdpListenPorts, common.JsonConfigs.ServerUdpListenPorts) {
	// 	r := intersection(newConfig.ServerUdpListenPorts, common.JsonConfigs.ServerUdpListenPorts)
	// 	add := subtraction(newConfig.ServerUdpListenPorts, r)
	// 	del := subtraction(common.JsonConfigs.ServerUdpListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Udp server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerUdpListenPorts, newConfig.ServerUdpListenPorts, del, add)
	// 		common.JsonConfigs.ServerUdpListenPorts = newConfig.ServerUdpListenPorts

	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			initUdpServer(fmt.Sprintf("UdpServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(100) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// // // reflush special udp server
	// // if !reflect.DeepEqual(newConfig.ServerUdpListenHosts, common.JsonConfigs.ServerUdpListenHosts) {
	// // 	r := intersection(newConfig.ServerUdpListenHosts, common.JsonConfigs.ServerUdpListenHosts)
	// // 	add := subtraction(newConfig.ServerUdpListenHosts, r)
	// // 	del := subtraction(common.JsonConfigs.ServerUdpListenHosts, r)

	// // 	if 0 != len(del) || 0 != len(add) {
	// // 		common.System("Udp server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerUdpListenHosts, newConfig.ServerUdpListenHosts, del, add)
	// // 		common.JsonConfigs.ServerUdpListenHosts = newConfig.ServerUdpListenHosts

	// // 		for _, port := range del {
	// // 			err = stopServer(port)
	// // 			if nil != err {
	// // 				common.Error("The %v file watcher stop [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// // 				return err
	// // 			}
	// // 		}

	// // 		for _, port := range add {
	// // 			listenAddr.Port = port
	// // 			initUdpServer(fmt.Sprintf("UdpServer-%v", port), listenAddr)
	// // 			time.Sleep(time.Duration(100) * time.Millisecond)
	// // 			err = startServer(port)
	// // 			if nil != err {
	// // 				common.Error("The %v file watcher start [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// // 				return err
	// // 			}
	// // 		}
	// // 	}
	// // }

	// // reflush http server
	// if !reflect.DeepEqual(newConfig.ServerHttpListenPorts, common.JsonConfigs.ServerHttpListenPorts) {
	// 	r := intersection(newConfig.ServerHttpListenPorts, common.JsonConfigs.ServerHttpListenPorts)
	// 	add := subtraction(newConfig.ServerHttpListenPorts, r)
	// 	del := subtraction(common.JsonConfigs.ServerHttpListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Http server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerHttpListenPorts, newConfig.ServerHttpListenPorts, del, add)
	// 		common.JsonConfigs.ServerHttpListenPorts = newConfig.ServerHttpListenPorts

	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [HttpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			initHttpServer(fmt.Sprintf("HttpServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(100) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [HttpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// // reflush https server
	// if !reflect.DeepEqual(newConfig.ServerHttpsListenPorts, common.JsonConfigs.ServerHttpsListenPorts) {
	// 	r := intersection(newConfig.ServerHttpsListenPorts, common.JsonConfigs.ServerHttpsListenPorts)
	// 	add := subtraction(newConfig.ServerHttpsListenPorts, r)
	// 	del := subtraction(common.JsonConfigs.ServerHttpsListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Https server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerHttpsListenPorts, newConfig.ServerHttpsListenPorts, del, add)
	// 		common.JsonConfigs.ServerHttpsListenPorts = newConfig.ServerHttpsListenPorts

	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [HttpsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			prepareCert()
	// 			initHttpsServer(fmt.Sprintf("HttpsServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(300) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [HttpsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// // reflush quic server
	// if !reflect.DeepEqual(newConfig.ServerQuicListenPorts, common.JsonConfigs.ServerQuicListenPorts) {
	// 	r := intersection(newConfig.ServerQuicListenPorts, common.JsonConfigs.ServerQuicListenPorts)
	// 	add := subtraction(newConfig.ServerQuicListenPorts, r)
	// 	del := subtraction(common.JsonConfigs.ServerQuicListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Quic server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerQuicListenPorts, newConfig.ServerQuicListenPorts, del, add)
	// 		common.JsonConfigs.ServerQuicListenPorts = newConfig.ServerQuicListenPorts
	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [QuicServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			initQuicServer(fmt.Sprintf("QuicServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(300) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [QuicServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// // reflush dns server
	// if !reflect.DeepEqual(newConfig.ServerDnsListenPorts, common.JsonConfigs.ServerDnsListenPorts) {
	// 	r := intersection(newConfig.ServerDnsListenPorts, common.JsonConfigs.ServerDnsListenPorts)
	// 	add := subtraction(newConfig.ServerDnsListenPorts, r)
	// 	del := subtraction(common.JsonConfigs.ServerDnsListenPorts, r)

	// 	if 0 != len(del) || 0 != len(add) {
	// 		common.System("Dns server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerDnsListenPorts, newConfig.ServerDnsListenPorts, del, add)
	// 		common.JsonConfigs.ServerDnsListenPorts = newConfig.ServerDnsListenPorts
	// 		for _, port := range del {
	// 			err = stopServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher stop [DnsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}

	// 		for _, port := range add {
	// 			listenAddr.Port = port
	// 			if 0 != len(common.JsonConfigs.ServerDnsListenPorts) {
	// 				saveDnsEntrys()
	// 			}
	// 			initDnsServer(fmt.Sprintf("DnsServer-%v", port), listenAddr)
	// 			time.Sleep(time.Duration(50) * time.Millisecond)
	// 			err = startServer(port)
	// 			if nil != err {
	// 				common.Error("The %v file watcher start [DnsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// if !reflect.DeepEqual(newConfig.ServerDnsAEntrys, common.JsonConfigs.ServerDnsAEntrys) ||
	// 	!reflect.DeepEqual(newConfig.ServerDns4AEntrys, common.JsonConfigs.ServerDns4AEntrys) {
	// 	common.System("Dns server entrys has changed, oldA=%v, old4A=%v, newA=%v, new4A=%v\n",
	// 		common.JsonConfigs.ServerDnsAEntrys, common.JsonConfigs.ServerDns4AEntrys, newConfig.ServerDnsAEntrys, newConfig.ServerDns4AEntrys)
	// 	common.JsonConfigs.ServerDnsAEntrys = newConfig.ServerDnsAEntrys
	// 	common.JsonConfigs.ServerDns4AEntrys = newConfig.ServerDns4AEntrys
	// 	saveDnsEntrys()
	// 	printDnsServerEntrys()
	// 	common.System("\n")
	// }

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

func intersection(aSet, bSet []uint16) []uint16 {
	r := make([]uint16, 0)
	for _, a := range aSet {
		for _, b := range bSet {
			if a == b {
				r = append(r, a)
			}
		}
	}

	return r
}

func subtraction(aSet, bSet []uint16) []uint16 {
	s := make([]uint16, 0)

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
