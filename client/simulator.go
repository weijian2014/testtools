package main

import (
	"../common"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	simulatorFullPath = ""
)

func sendIosByHttp() {
	if !simulatorIsExist() {
		panic("simulator executable file does not exist")
	}

	ret, err := common.Command(simulatorFullPath, "--ios", "-s", common.Configs.ClientBindIpAddress, "-d", sendToServerIpAddress,
		"--dport", strconv.Itoa(int(sentToServerPort)))
	if nil != err {
		panic(err)
	}

	if !strings.Contains(string(ret), "successful") {
		panic("IOS simulator send packet failed")
	}

	//fmt.Printf("execute info:%v\n", string(ret))
	all := strings.Split(string(ret), "=")
	scrPortLine := strings.Split(all[3], ",")
	srcPort := scrPortLine[0]
	fmt.Printf("IOS simulator[%v:%v]----Http server[http://%v:%v] ok\n", common.Configs.ClientBindIpAddress, srcPort, sendToServerIpAddress, sentToServerPort)
}

func sendWindowsByHttp() {
	if !simulatorIsExist() {
		panic("simulator executable file does not exist")
	}

	ret, err := common.Command(simulatorFullPath, "--win", "-s", common.Configs.ClientBindIpAddress, "-d", sendToServerIpAddress,
		"--dport", strconv.Itoa(int(sentToServerPort)))
	if nil != err {
		panic(err)
	}

	if !strings.Contains(string(ret), "successful") {
		panic("IOS simulator send packet failed")
	}

	//fmt.Printf("execute info:%v\n", string(ret))
	all := strings.Split(string(ret), "=")
	scrPortLine := strings.Split(all[3], ",")
	srcPort := scrPortLine[0]
	fmt.Printf("Win simulator[%v:%v]----Http server[http://%v:%v] ok\n", common.Configs.ClientBindIpAddress, srcPort, sendToServerIpAddress, sentToServerPort)
}

func simulatorIsExist() bool {
	simulatorFullPath = common.CurrDir + "/../tools/simulator/bin/simulator"
	_, err := os.Stat(simulatorFullPath)
	if nil == err || os.IsExist(err) {
		err = os.Chmod(simulatorFullPath, 0777)
		if nil != err {
			panic(err)
		}
		return true
	}

	simulatorFullPath = common.CurrDir + "/simulator"
	_, err = os.Stat(simulatorFullPath)
	if nil == err || os.IsExist(err) {
		err = os.Chmod(simulatorFullPath, 0777)
		if nil != err {
			panic(err)
		}
		return true
	}

	simulatorFullPath = common.CurrDir + "/tools/simulator/bin/simulator"
	_, err = os.Stat(simulatorFullPath)
	if nil == err || os.IsExist(err) {
		err = os.Chmod(simulatorFullPath, 0777)
		if nil != err {
			panic(err)
		}
		return true
	}

	return false
}
