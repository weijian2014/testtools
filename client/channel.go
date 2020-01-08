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

func sendHttpByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doHttp(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send http by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func sendHttpsByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doHttps(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send https by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func sendGQuicByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doGQuic(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send gquic by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func sendIQuicByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doIQuic(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send iquic by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
}

func sendDnsByRange() {
	start := time.Now().Unix()
	preChannel()

	var i uint64
	for i = 0; i < common.JsonConfigs.ClientRangeModeChannelNumber; i++ {
		go doDns(i)
	}

	wg.Wait()
	end := time.Now().Unix()
	common.Fatal("Send dns by range done, start timestamp %v, end timestamp %v, time elapse %v \n", start, end, (end - start))
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

	close(ch)
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

	close(ch)
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

	close(ch)
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

	close(ch)
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

	close(ch)
}
