# 一个 Golang 开发的 TCP, UDP, HTTP 1.0/2.0/3.0, HTTPS, QUIC, DNS 协议的 echo 测试工具

- 分 client 和 server 端, 使用 json 配置文件, 支持配置热更新
- client 端支持指定源地址,目标端口和地址
- HTTP 和 HTTPS 支持文件下载和上传

# 使用

```
# testtools -h
Usage of /root/testtools/testtools:
  -alpn string
    	The ALPN of QUIC protocol, which example as "aaa,bbb,ccc", "ietf-quic-v1-echo-example" is hard code both in server and client
  -d string
    	The destination IP address of client
    	This parameter takes precedence over ClientSendToIpv4Address or ClientSendToIpv6Address in the config.json file

  -debug int
    	The client log level, 0-Debug, 1-Info, 2-System, 3-Warn, 4-Error, 5-Fatal
  -dns
    	Using DNS protocol
  -dport uint
    	The port of server, valid only for UDP, TCP, QUIC protocols
  -f string
    	The path of config.json file, support for absolute and relative paths (default "/root/weijian/testtools/config.json")
  -h	Show help
  -http
    	Using HTTP protocol
  -http2
    	Using HTTP2 protocol
  -http3
    	Using HTTP3 over QUIC protocol
  -https
    	Using HTTPS protocol
  -n uint
    	The number of client send data to server, valid only for UDP, TCP, QUIC protocols (default 1)
  -quic
    	Using QUIC protocol
  -s string
    	The source IP address of client
    	This parameter takes precedence over clientBindIpAddress in the config.json file
    	If the parameter is an IPv6 address, the client will send data to the ClientSendToIpv6Address of config.json file
  -server
    	As server running, default as client
  -sni string
    	The SNI of QUIC protocol
  -t uint
    	The timeout seconds of client read or write
  -tcp
    	Using TCP protocol
  -udp
    	Using UDP protocol
  -w uint
    	The second waiting to send before, support TCP, UDP, QUIC and DNS protocol
```

# go mod 命令说明

    go mod download: 下载依赖的 module 到本地 cache
    go mod edit: 编辑 go.mod
    go mod graph: 打印模块依赖图
    go mod init: 在当前目录下初始化 go.mod(就是会新建一个 go.mod 文件)
    go mod tidy: 整理依赖关系，会添加丢失的 module，删除不需要的 module
    go mod vender: 将依赖复制到 vendor 下
    go mod verify: 校验依赖
    go mod why: 解释为什么需要依赖
    go list -m -json all: JSON格式显示所有Import库信息

# GoLand 配置

    settings---GO---Go Modules(vgo)---勾选 Enable Go Modules(vgo) integration  #开启Go Modules
    settings---GO---Go Modules(vgo)---proxy中填入https://goproxy.io            #开启代理

# 项目中使用 Go Modules, 并构建项目

    mkdir -p /opt/go/gopath
    echo "export GOROOT=/opt/go" >> ~/.profile
    echo "export PATH=${PATH}:${GOROOT}/bin" >> ~/.profile
    echo "export GOPATH=${GOROOT}/gopath" >> ~/.profile
    echo "export GO111MODULE=on" >> ~/.profile
    echo "export GOPROXY=https://goproxy.cn" >> ~/.profile

    source ~/.profile
    cd testtools
    go mod init testtools
    go mod tidy
    gofmt -l -w .
    go build .

# 格式化整个项目

    gofmt -l -w .

# go.mod 中加入 quic-go 的 google 版本， 0810051 为 google 版本的最后一个 commit id

    require github.com/quic-go/quic-go 0810051

# Go Build Environment:

    GOPROXY=https://goproxy.io;GOOS=linux;GOARCH=amd64;GO111MODULE=on
