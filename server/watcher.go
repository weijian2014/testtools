package server

import (
	"fmt"
	"reflect"
	"testtools/common"
	"time"

	"github.com/fsnotify/fsnotify"
)

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
				if !ok {
					common.System("The %v file no ok in evnet!", common.FlagInfos.ConfigFileFullPath)
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					err := reflushServers()
					if nil != err {
						common.System("The %v file watcher reflush servers fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					common.System("The %v file no ok in error!", common.FlagInfos.ConfigFileFullPath)
					return
				}

				if nil != err {
					common.System("The %v file watcher get error, err: %v\n", common.FlagInfos.ConfigFileFullPath, err)
					return
				}
			}
		}
	}()

	common.System("The %v file watcher start ok", common.FlagInfos.ConfigFileFullPath)

	err = watcher.Add(common.FlagInfos.ConfigFileFullPath)
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

	common.JsonConfigs.CommonLogLevel = newConfig.CommonLogLevel
	common.JsonConfigs.CommonLogRoll = newConfig.CommonLogRoll
	common.JsonConfigs.CommonRecvBufferSizeBytes = newConfig.CommonRecvBufferSizeBytes
	common.JsonConfigs.ServerCounterOutputIntervalSeconds = newConfig.ServerCounterOutputIntervalSeconds
	common.JsonConfigs.ServerListenHost = newConfig.ServerListenHost
	common.JsonConfigs.ServerDnsAEntrys = newConfig.ServerDnsAEntrys
	common.JsonConfigs.ServerDns4AEntrys = newConfig.ServerDns4AEntrys
	common.JsonConfigs.ServerSendData = newConfig.ServerSendData
	common.JsonConfigs.ClientBindIpAddress = newConfig.ClientBindIpAddress
	common.JsonConfigs.ClientSendToIpv4Address = newConfig.ClientSendToIpv4Address
	common.JsonConfigs.ClientSendToIpv6Address = newConfig.ClientSendToIpv6Address
	common.JsonConfigs.ClientSendData = newConfig.ClientSendData

	listenAddr := common.IpAndPort{Ip: common.JsonConfigs.ServerListenHost, Port: 0}

	// reflush tcp server
	if !reflect.DeepEqual(newConfig.ServerTcpListenPorts, common.JsonConfigs.ServerTcpListenPorts) {
		r := intersection(newConfig.ServerTcpListenPorts, common.JsonConfigs.ServerTcpListenPorts)
		add := subtraction(newConfig.ServerTcpListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerTcpListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Tcp server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerTcpListenPorts, newConfig.ServerTcpListenPorts, del, add)
			common.JsonConfigs.ServerTcpListenPorts = newConfig.ServerTcpListenPorts

			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [TcpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				initTcpServer(fmt.Sprintf("TcpServer-%v", port), listenAddr)
				time.Sleep(time.Duration(200) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [TcpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	// reflush udp server
	if !reflect.DeepEqual(newConfig.ServerUdpListenPorts, common.JsonConfigs.ServerUdpListenPorts) {
		r := intersection(newConfig.ServerUdpListenPorts, common.JsonConfigs.ServerUdpListenPorts)
		add := subtraction(newConfig.ServerUdpListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerUdpListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Udp server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerUdpListenPorts, newConfig.ServerUdpListenPorts, del, add)
			common.JsonConfigs.ServerUdpListenPorts = newConfig.ServerUdpListenPorts

			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				initUdpServer(fmt.Sprintf("UdpServer-%v", port), listenAddr)
				time.Sleep(time.Duration(100) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [UdpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	// reflush http server
	if !reflect.DeepEqual(newConfig.ServerHttpListenPorts, common.JsonConfigs.ServerHttpListenPorts) {
		r := intersection(newConfig.ServerHttpListenPorts, common.JsonConfigs.ServerHttpListenPorts)
		add := subtraction(newConfig.ServerHttpListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerHttpListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Http server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerHttpListenPorts, newConfig.ServerHttpListenPorts, del, add)
			common.JsonConfigs.ServerHttpListenPorts = newConfig.ServerHttpListenPorts

			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [HttpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				initHttpServer(fmt.Sprintf("HttpServer-%v", port), listenAddr)
				time.Sleep(time.Duration(100) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [HttpServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	// reflush https server
	if !reflect.DeepEqual(newConfig.ServerHttpsListenPorts, common.JsonConfigs.ServerHttpsListenPorts) {
		r := intersection(newConfig.ServerHttpsListenPorts, common.JsonConfigs.ServerHttpsListenPorts)
		add := subtraction(newConfig.ServerHttpsListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerHttpsListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Https server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerHttpsListenPorts, newConfig.ServerHttpsListenPorts, del, add)
			common.JsonConfigs.ServerHttpsListenPorts = newConfig.ServerHttpsListenPorts

			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [HttpsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				prepareCert()
				initHttpsServer(fmt.Sprintf("HttpsServer-%v", port), listenAddr)
				time.Sleep(time.Duration(300) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [HttpsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	// reflush quic server
	if !reflect.DeepEqual(newConfig.ServerQuicListenPorts, common.JsonConfigs.ServerQuicListenPorts) {
		r := intersection(newConfig.ServerQuicListenPorts, common.JsonConfigs.ServerQuicListenPorts)
		add := subtraction(newConfig.ServerQuicListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerQuicListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Quic server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerQuicListenPorts, newConfig.ServerQuicListenPorts, del, add)
			common.JsonConfigs.ServerQuicListenPorts = newConfig.ServerQuicListenPorts
			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [QuicServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				initQuicServer(fmt.Sprintf("QuicServer-%v", port), listenAddr)
				time.Sleep(time.Duration(300) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [QuicServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	// reflush dns server
	if !reflect.DeepEqual(newConfig.ServerDnsListenPorts, common.JsonConfigs.ServerDnsListenPorts) {
		r := intersection(newConfig.ServerDnsListenPorts, common.JsonConfigs.ServerDnsListenPorts)
		add := subtraction(newConfig.ServerDnsListenPorts, r)
		del := subtraction(common.JsonConfigs.ServerDnsListenPorts, r)

		if 0 != len(add) && 0 != len(add) {
			common.System("Dns server port has changed, old=%v, new=%v, del=%v, add=%v\n", common.JsonConfigs.ServerDnsListenPorts, newConfig.ServerDnsListenPorts, del, add)
			common.JsonConfigs.ServerDnsListenPorts = newConfig.ServerDnsListenPorts
			for _, port := range del {
				err = common.StopServer(port)
				if nil != err {
					common.System("The %v file watcher stop [DnsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}

			for _, port := range add {
				listenAddr.Port = port
				if 0 != len(common.JsonConfigs.ServerDnsListenPorts) {
					saveDnsEntrys()
				}
				initDnsServer(fmt.Sprintf("DnsServer-%v", port), listenAddr)
				time.Sleep(time.Duration(50) * time.Millisecond)
				err = common.StartServer(port)
				if nil != err {
					common.System("The %v file watcher start [DnsServer-%v] server fail, err: %v\n", common.FlagInfos.ConfigFileFullPath, port, err)
				}
			}
		}
	}

	common.JsonConfigs = newConfig
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
