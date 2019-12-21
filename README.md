# go mod命令说明
    go mod download: 下载依赖的 module 到本地 cache
    go mod edit: 编辑 go.mod
    go mod graph: 打印模块依赖图
    go mod init: 在当前目录下初始化 go.mod(就是会新建一个 go.mod 文件)
    go mod tidy: 整理依赖关系，会添加丢失的 module，删除不需要的 module
    go mod vender: 将依赖复制到 vendor 下
    go mod verify: 校验依赖
    go mod why: 解释为什么需要依赖

# GoLand配置
    settings---GO---Go Modules(vgo)---勾选 Enable Go Modules(vgo) integration  #开启Go Modules
    settings---GO---Go Modules(vgo)---proxy中填入https://goproxy.io            #开启代理

# 项目中使用Go Modules, 并构建项目
    mkdir -p /opt/go/gopath
    echo "export PATH=${PATH}:/opt/go" >> ~/.profile
    echo "export GOROOT=/opt/go" >> ~/.profile
    echo "export GOPATH=/opt/go/gopath" >> ~/.profile
    echo "export GOPROXY=https://goproxy.io" >> ~/.profile
    echo "export GO111MODULE=on" >> ~/.profile
    source ~/.profile
    cd testtools
    go mod init testtools
    go mod tidy
    cd server
    go build .
    cd ../client
    go build .

# go.mod中加入quic-go的google版本， 0810051为google版本的最后一个commit id
    require github.com/lucas-clemente/quic-go 0810051

# Go Build Environment:
    GOPROXY=https://goproxy.io;GOOS=linux;GOARCH=amd64;GO111MODULE=on