package client

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testtools/common"
	"time"
)

var (
	channels       []chan string
	totalSendCount uint64 = 0
	sleepTime      int64  = 5
	undoneChannels int64  = 0
)

func sendByRange(protocolType int) {
	undoneChannels = int64(common.JsonConfigs.ClientRangeModeChannelNumber)
	start := time.Now().Unix()
	preChannel()

	var i uint64
	switch protocolType {
	case common.TcpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doTcp(i)
		}
	case common.UdpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doUdp(i)
		}
	case common.HttpProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doHttp(i)
		}
	case common.HttpsProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doHttps(i)
		}
	case common.GQuicProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doGQuic(i)
		}
	case common.IQuicProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doIQuic(i)
		}
	case common.DnsProtocolType:
		for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
			go doDns(i)
		}
	default:
		common.Fatal("Unknown protocol type %v\n", protocolType)
	}

	for {
		if 0 != atomic.LoadInt64(&undoneChannels) {
			var total uint64 = totalSendCount
			completed, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(total)/float32(clientBindIpAddressRangeLength)*100), 64)
			if nil != err {
				panic(err)
			}
			end := time.Now().Unix()
			diff := end - start
			if 0 == diff {
				time.Sleep(time.Duration(sleepTime) * time.Second)
				continue
			}

			common.Error("Doing...(interval %v second)\n\tsend count: %v\n\tunsend count: %v\n\tchannel count: %v\n\tprogress rate: %v%%\n\tsend count per second: %v\n\ttime elapse(second): %v\n",
				sleepTime, total, clientBindIpAddressRangeLength - total, common.JsonConfigs.ClientRangeModeChannelNumber, completed, total/uint64(diff), diff)
			time.Sleep(time.Duration(sleepTime) * time.Second)
			continue
		} else {
			break
		}
	}

	end := time.Now().Unix()
	common.Error("Done\n\tstart timestamp: %v\n\tend timestamp: %v\n\ttime elapse(second): %v\n\tchannel count: %v\n\tclient ip count: %v\n\ttotal send count: %v\n",
		start, end, end-start, common.JsonConfigs.ClientRangeModeChannelNumber, clientBindIpAddressRangeLength, totalSendCount)
}

func preChannel() {
	var channelBufferSize uint64 = (clientBindIpAddressRangeLength / common.JsonConfigs.ClientRangeModeChannelNumber) + 3

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		channels = append(channels, make(chan string, channelBufferSize))
	}

	var sendCount uint64 = 0
	var channelIndex uint64 = 0
	for _, bindIp := range clientBindIpAddressRange {
		sendCount++
		channelIndex = sendCount % common.JsonConfigs.ClientRangeModeChannelNumber
		channels[channelIndex] <- bindIp
	}

	for _, ch := range channels {
		ch <- "end"
	}
}

func doTcp(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByTcp(ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doUdp(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByUdp(ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doHttp(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByHttp(ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doHttps(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByHttps(ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doGQuic(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByGQuic("gQuic", ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doIQuic(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByGQuic("iQuic", ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}

func doDns(index uint64) {
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByDns(ip)
		atomic.AddUint64(&totalSendCount, 1)
	}

	atomic.AddInt64(&undoneChannels, -1)
	defer close(ch)
}
