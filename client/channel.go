package client

import (
	"fmt"
	"sync"
	"testtools/common"
)

var (
	channels []chan string
	wg       sync.WaitGroup
)

func sendTcpByRange() {
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doTcp(i)
	}

	wg.Wait()
	fmt.Printf("Send tcp by range done\n")
}

func sendUdpByRange() {
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doUdp(i)
	}

	wg.Wait()
	fmt.Printf("Send tcp by range done\n")
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