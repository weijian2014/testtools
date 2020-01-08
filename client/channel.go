package client

import (
	"sync"
	"sync/atomic"
	"testtools/common"
	"time"
)

var (
	channels       []chan string
	wg             *sync.WaitGroup
	totalSendTimes uint64 = 0
)

func sendByRange(protocolType int) {
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

	time.Sleep(time.Duration(3) * time.Second)
	wg.Wait()
	end := time.Now().Unix()
	common.Error("Send by range done\n\tstart timestamp: %v\n\tend timestamp: %v\n\ttime elapse: %v\n\tchannel number: %v\n\tclient ip number: %v\n\ttotal send times: %v\n",
		start, end, (end - start), common.JsonConfigs.ClientRangeModeChannelNumber, len(clientBindIpAddressRange), totalSendTimes)
}

func preChannel() {
	wg = &sync.WaitGroup{}
	var channelBufferSize uint64 = (uint64(len(clientBindIpAddressRange)) / common.JsonConfigs.ClientRangeModeChannelNumber) + 3

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		channels = append(channels, make(chan string, channelBufferSize))
	}

	var sendTimes uint64 = 0
	var channelIndex uint64 = 0
	for _, bindIp := range clientBindIpAddressRange {
		sendTimes++
		channelIndex = sendTimes % common.JsonConfigs.ClientRangeModeChannelNumber
		channels[channelIndex] <- bindIp
	}

	for _, ch := range channels {
		ch <- "end"
	}
}

func doTcp(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByTcp(ip)
		atomic.AddUint64(&totalSendTimes, 1)
	}

	defer close(ch)
}

func doUdp(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByUdp(ip)
	}

	defer close(ch)
}

func doHttp(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByHttp(ip)
	}

	defer close(ch)
}

func doHttps(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByHttps(ip)
	}

	defer close(ch)
}

func doGQuic(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByGQuic("gQuic", ip)
	}

	defer close(ch)
}

func doIQuic(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByGQuic("iQuic", ip)
	}

	defer close(ch)
}

func doDns(index uint64) {
	defer wg.Done()
	wg.Add(1)
	ch := channels[index]
	for {
		ip := <-ch
		if "end" == ip {
			break
		}

		sendByDns(ip)
	}

	defer close(ch)
}
