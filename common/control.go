package common

import (
	"fmt"
	"sync"
	"time"
)

const (
	StartServerControlOption int = iota // StartServer=0
	StopServerControlOption
	MaxControlOption
)

var (
	controlChannelsMap      map[uint16]chan int
	controlChannelsMapGuard sync.Mutex
)

func init() {
	controlChannelsMap = make(map[uint16]chan int)
}

func InsertControlChannel(port uint16, ctrlChan chan int) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	_, ok := controlChannelsMap[port]
	if ok {
		return fmt.Errorf("The control channel already exists")
	}

	controlChannelsMap[port] = ctrlChan
	return nil
}

func DeleteControlChannel(port uint16) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	_, ok := controlChannelsMap[port]
	if !ok {
		return fmt.Errorf("The control channel does not exist")
	}

	delete(controlChannelsMap, port)
	return nil
}

func SendOptionToControlChannel(port uint16, option int) error {
	if 65535 <= port {
		return fmt.Errorf("Invalid port")
	}

	err := checkControlOption(option)
	if nil != err {
		return err
	}

	controlChannelsMapGuard.Lock()
	defer controlChannelsMapGuard.Unlock()

	ctrlChan, ok := controlChannelsMap[port]
	if !ok {
		return fmt.Errorf("The control channel does not exist")
	}

	ctrlChan <- option
	return nil
}

func StartAllServers() error {
	for port, _ := range controlChannelsMap {
		err := SendOptionToControlChannel(port, StartServerControlOption)
		if nil != err {
			return err
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return nil
}

func StopServer(port uint16) error {
	return SendOptionToControlChannel(port, StopServerControlOption)
}

func checkControlOption(option int) error {
	if StartServerControlOption > option || MaxControlOption <= option {
		return fmt.Errorf("Invalid option")
	}

	return nil
}
