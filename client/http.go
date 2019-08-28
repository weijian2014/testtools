package main

import (
	"../common"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func sendByHttp() {
	localAddr, err := net.ResolveIPAddr("ip", common.Configs.ClientBindIpAddress)
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
		reqeustUrl = fmt.Sprintf("http://%s:%d", sendToServerIpAddress, sentToServerPort)
	} else {
		reqeustUrl = fmt.Sprintf("http://[%s]:%d", sendToServerIpAddress, sentToServerPort)
	}

	fmt.Printf("Http client bind on %v, will reqeust to %v\n", common.Configs.ClientBindIpAddress, reqeustUrl)
	for i := 1; i <= clientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Http client new request failed, times[%d], err : %v\n", i, err.Error())
			return
		}
		req.Header.Add("ClientSendData", common.Configs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Http client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			fmt.Printf("Http client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		fmt.Printf("Http client[%v]----Http server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, common.Configs.ClientSendData, body)
	}
}

func sendByHttps() {
	localAddr, err := net.ResolveIPAddr("ip", common.Configs.ClientBindIpAddress)
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
		reqeustUrl = fmt.Sprintf("https://%s:%d", sendToServerIpAddress, sentToServerPort)
	} else {
		reqeustUrl = fmt.Sprintf("https://[%s]:%d", sendToServerIpAddress, sentToServerPort)
	}

	fmt.Printf("Https client bind on %v, will reqeust to %v\n", common.Configs.ClientBindIpAddress, reqeustUrl)
	for i := 1; i <= clientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.Configs.ClientSendData))
		if err != nil {
			fmt.Printf("Https client new request failed, times[%d], err : %v\n", i, err.Error())
			return
		}
		req.Header.Add("ClientSendData", common.Configs.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Https client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			fmt.Printf("Https client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			return
		}

		fmt.Printf("Https client[%v]----Https server[%v], times[%d]:\n\tsend: %s\n\trecv: %s", localTCPAddr.String(), req.Host, i, common.Configs.ClientSendData, body)
	}
}
