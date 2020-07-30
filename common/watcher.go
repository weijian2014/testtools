package common

import (
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
)

func StartConfigFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	configFileDir := filepath.Dir(FlagInfos.ConfigFileFullPath)
	err = watcher.Add(configFileDir)
	if err != nil {
		panic(err)
	}

	System("The %v file watcher start ok", FlagInfos.ConfigFileFullPath)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					System("The %v file content has changed!", FlagInfos.ConfigFileFullPath)
					getConfigFileDiff()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				if nil != err {
					System("The %v file watcher error, err: %v\n", FlagInfos.ConfigFileFullPath, err)
					return
				}
			}
		}
	}()
}

func getConfigFileDiff() ([]uint16, []uint16, error) {
	newConfig, err := LoadConfigFile(FlagInfos.ConfigFileFullPath)
	if nil != err {
		return nil, nil, err
	}

	add := make([]uint16, 0)
	del := make([]uint16, 0)

	// tcp
	if !reflect.DeepEqual(newConfig.ServerTcpListenPorts, JsonConfigs.ServerTcpListenPorts) {
		r := intersection(newConfig.ServerTcpListenPorts, JsonConfigs.ServerTcpListenPorts)
		add = append(add, subtraction(newConfig.ServerTcpListenPorts, r)...)
		del = append(del, subtraction(JsonConfigs.ServerTcpListenPorts, r)...)

		System("add=[%v], del[%v]\n", add, del)
		return add, del, nil

		// if 0 != len(r) {
		// 	for _, port := range newConfig.ServerTcpListenPorts {
		// 		err = StopServer(port)
		// 		if nil != err {
		// 			return nil, nil, err
		// 		}
		// 	}
		// }
	}

	return nil, nil, nil
}

func intersection(aSet, bSet []uint16) []uint16 {
	r := make([]uint16, 0)
	for _, a := range aSet {
		for _, b := range bSet {
			if a == b {
				r = append(r, a)
			}
		}
	}

	return r
}

func subtraction(aSet, bSet []uint16) []uint16 {
	s := aSet
	for ai, a := range aSet {
		for _, b := range bSet {
			if a == b {
				s = append(s[:ai], s[ai+1:]...)
			}
		}
	}

	return s
}
