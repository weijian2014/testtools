package client

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testtools/common"
	"time"
)

func sendByHttp(localIp string) {
	localAddr, err := net.ResolveIPAddr("ip", localIp)
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var reqeustUrl string
	if true == strings.Contains(sendToServerIpAddress, ".") {
		reqeustUrl = fmt.Sprintf("http://%s:%d", sendToServerIpAddress, common.FlagInfos.SentToServerPort)
	} else {
		reqeustUrl = fmt.Sprintf("http://[%s]:%d", sendToServerIpAddress, common.FlagInfos.SentToServerPort)
	}

	common.Info("Http client bind on %v, will reqeust to %v\n", localIp, reqeustUrl)

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.JsonConfigs.ClientSendData))
		if err != nil {
			common.Warn("Http client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.JsonConfigs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			common.Warn("Http client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			common.Warn("Http client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		common.Info("Http client[%v]----Http server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, common.JsonConfigs.ClientSendData, body)
	}
}

func sendByHttps(localIp string) {
	localAddr, err := net.ResolveIPAddr("ip", localIp)
	if err != nil {
		panic(err)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略证书
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var reqeustUrl string
	if true == strings.Contains(sendToServerIpAddress, ".") {
		reqeustUrl = fmt.Sprintf("https://%s:%d", sendToServerIpAddress, common.FlagInfos.SentToServerPort)
	} else {
		reqeustUrl = fmt.Sprintf("https://[%s]:%d", sendToServerIpAddress, common.FlagInfos.SentToServerPort)
	}

	common.Info("Https client bind on %v, will reqeust to %v\n", localIp, reqeustUrl)

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.JsonConfigs.ClientSendData))
		if err != nil {
			common.Warn("Https client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.JsonConfigs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			common.Warn("Https client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			common.Warn("Https client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		common.Info("Https client[%v]----Https server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, common.JsonConfigs.ClientSendData, body)
	}
}
