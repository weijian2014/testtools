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

	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
)

func sendByHttp10(localAddr, remoteAddr *common.IpAndPort) {
	lAddr, err := net.ResolveTCPAddr("tcp", localAddr.String())
	if nil != err {
		panic(err)
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: lAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	reqeustUrl := fmt.Sprintf("http://%v", remoteAddr.String())

	common.Info("Http client bind on %v, will reqeust to %v\n", localAddr.String(), reqeustUrl)

	var i uint64
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		client := &http.Client{Transport: tr}
		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Http client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.FlagInfos.ClientSendData)
		req.Proto = "HTTP/1.0"

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

		common.Info("Http client[%v]----Http server[%v], times[%d]:\n\tresp: proto=%v, status=%v, contentLength=%v, header=%v\n\tsend: %s\n\trecv: %s",
			localAddr.String(), req.Host, i, resp.Proto, resp.Status, resp.ContentLength, resp.Header, common.FlagInfos.ClientSendData, body)
	}
}

func sendByHttp11(localAddr, remoteAddr *common.IpAndPort, isEnableHttp2 bool) {
	lAddr, err := net.ResolveTCPAddr("tcp", localAddr.String())
	if nil != err {
		panic(err)
	}

	reqeustUrl := fmt.Sprintf("https://%v", remoteAddr.String())

	common.Info("Https client bind on %v, will reqeust to %v\n", localAddr.String(), reqeustUrl)

	var i uint64
	var client *http.Client
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 忽略证书
				},
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					LocalAddr: lAddr,
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}

		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Https client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.FlagInfos.ClientSendData)
		req.Proto = "HTTP/1.1"

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

		common.Info("Https client[%v]----Https server[%v], times[%d]:\n\tresp: proto=%v, status=%v, contentLength=%v, header=%v\n\tsend: %s\n\trecv: %s",
			localAddr.String(), req.Host, i, resp.Proto, resp.Status, resp.ContentLength, resp.Header, common.FlagInfos.ClientSendData, body)
	}
}

func sendByHttp20(localAddr, remoteAddr *common.IpAndPort) {
	reqeustUrl := fmt.Sprintf("https://%v", remoteAddr.String())
	common.Info("Https client bind on %v, will reqeust to %v\n", localAddr.String(), reqeustUrl)

	var i uint64
	var client *http.Client
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		// 启动HTTP/2协议
		client = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true, // Skip TLS dial
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 忽略证书
				},
			},
		}

		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Https client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.FlagInfos.ClientSendData)
		req.Proto = "HTTP/2.0"

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

		common.Info("Https client[%v]----Https server[%v], times[%d]:\n\tresp: proto=%v, status=%v, contentLength=%v, header=%v\n\tsend: %s\n\trecv: %s",
			localAddr.String(), req.Host, i, resp.Proto, resp.Status, resp.ContentLength, resp.Header, common.FlagInfos.ClientSendData, body)
	}
}

func sendByHttp30(localAddr, remoteAddr *common.IpAndPort) {
	reqeustUrl := fmt.Sprintf("https://%v", remoteAddr.String())
	common.Info("Http3 client bind on %v, will reqeust to %v\n", localAddr.String(), reqeustUrl)

	var i uint64
	var client *http.Client
	for i = 1; i <= common.FlagInfos.ClientSendNumbers; i++ {
		// send request
		client = &http.Client{
			Transport: &http3.RoundTripper{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 忽略证书
				},
			},
		}

		req, err := http.NewRequest("GET", reqeustUrl, strings.NewReader(common.FlagInfos.ClientSendData))
		if err != nil {
			common.Warn("Http3 client new request failed, times[%d], err : %v\n", i, err.Error())
			continue
		}
		req.Header.Add("ClientSendData", common.FlagInfos.ClientSendData)

		resp, err := client.Do(req)
		if err != nil {
			common.Warn("Http3 client request to %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		// receive response
		body, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			common.Warn("Http3 client response from %v failed, times[%d], err : %v\n", reqeustUrl, i, err.Error())
			continue
		}

		common.Info("Http3 client[%v]----Http3 server[%v], times[%d]:\n\tresp: proto=%v, status=%v, contentLength=%v, header=%v\n\tsend: %s\n\trecv: %s",
			localAddr.String(), req.Host, i, resp.Proto, resp.Status, resp.ContentLength, resp.Header, common.FlagInfos.ClientSendData, body)
	}
}
