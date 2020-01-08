package client

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testtools/common"
	"time"
)

var (
	sleepSeconds     int64  = 5
	undoneGoroutines int64  = 0
	totalSendCount   uint64 = 0
)

func sendByRange(protocolType int) {
	undoneGoroutines = int64(common.JsonConfigs.ClientRangeModeThreadNumber)
	start := time.Now().Unix()

	var i uint64
	switch protocolType {
	case common.TcpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doTcp(i)
		}
	case common.UdpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doUdp(i)
		}
	case common.HttpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doHttp(i)
		}
	case common.HttpsProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doHttps(i)
		}
	case common.GQuicProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doGQuic(i)
		}
	case common.IQuicProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doIQuic(i)
		}
	case common.DnsProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeThreadNumber; i++ {
			go doDns(i)
		}
	default:
		common.Fatal("Unknown protocol type %v\n", protocolType)
	}

	protocolName := common.ProtocolToString(protocolType)
	for {
		if 0 != atomic.LoadInt64(&undoneGoroutines) {
			var totalSend uint64 = totalSendCount
			completed, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(totalSend)/float32(clientBindIpAddressRangeLength)*100), 64)
			if nil != err {
				panic(err)
			}
			end := time.Now().Unix()
			diff := end - start
			if 0 == diff {
				continue
			}

			common.Error("%v doing...(interval %v second)\n\tthread count: %v\n\tsend count: %v\n\tunsend count: %v\n\tprogress rate: %v%%\n\ttime elapse(second): %v\n\tsend count per second: %v\n",
				protocolName, sleepSeconds, common.JsonConfigs.ClientRangeModeThreadNumber, totalSend, clientBindIpAddressRangeLength-totalSend, completed, diff, totalSend/uint64(diff))
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
			continue
		} else {
			break
		}
	}

	end := time.Now().Unix()
	diff := end - start
	if 0 == diff {
		common.Error("%v done:\n\tthread count: %v\n\tstart timestamp: %v\n\tend timestamp: %v\n\ttime elapse(second): %v\n\tclient ip count: %v\n\ttotal send count: %v\n",
			protocolName, common.JsonConfigs.ClientRangeModeThreadNumber, start, end, diff, clientBindIpAddressRangeLength, totalSendCount)
	} else {
		common.Error("%v done:\n\tthread count: %v\n\tstart timestamp: %v\n\tend timestamp: %v\n\ttime elapse(second): %v\n\tclient ip count: %v\n\ttotal send count: %v\n\tsend count per second: %v\n",
			protocolName, common.JsonConfigs.ClientRangeModeThreadNumber, start, end, diff, clientBindIpAddressRangeLength, totalSendCount, totalSendCount/uint64(diff))
	}

}

func doTcp(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByTcp(ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doUdp(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByUdp(ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doHttp(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByHttp(ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doHttps(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByHttps(ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doGQuic(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByGQuic("gQuic", ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doIQuic(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByGQuic("iQuic", ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}

func doDns(index uint64) {
	var i uint64
	for i = index; i < clientBindIpAddressRangeLength; i++ {
		ip := clientBindIpAddressRange[i]
		sendByDns(ip)
		i += common.JsonConfigs.ClientRangeModeThreadNumber
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneGoroutines, -1)
}
