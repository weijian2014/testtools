package client

import (
	"sync"
	"testtools/common"
	"time"
)

var (
	channels []chan string
	wg       sync.WaitGroup
)

func sendTcpByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doTcp(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send tcp by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func sendUdpByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doUdp(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send udp by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func preChannel() {
	var channelBufferSize uint64 = (uint64(len(clientBindIpAddressRange)) / common.JsonConfigs.ClientRangeModeChannelNumber) + 2

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
	}

	close(ch)
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

	close(ch)
}
