package server

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testtools/common"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
)

func initHttpServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func() {
		common.Debug("%v server control coroutine running...\n", serverName)
		server := newServer(listenAddr)

		c := make(chan int)
		err := insertControlChannel(listenAddr.String(), c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case StartServerControlOption:
				common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
				go server.ListenAndServe()
				isExit = false
			case StopServerControlOption:
				common.System("%v server stop\n", serverName)
				server.SetKeepAlivesEnabled(false)
				server.Shutdown(context.Background())
				err = deleteControlChannel(listenAddr.String())
				if nil != err {
					common.Error("Delete control channel fial, erro: %v", err)
				}
				isExit = true
			default:
				isExit = false
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}()
}

// support HTTP1.1 and HTTP2.0
func initHttpsServer(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func(serverName string, listenAddr common.IpAndPort) {
		common.Debug("%v server control coroutine running...\n", serverName)
		server := newServer(listenAddr)

		c := make(chan int)
		err := insertControlChannel(listenAddr.String(), c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case StartServerControlOption:
				{
					common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
					go server.ListenAndServeTLS(certificateFileFullPath, privateKeyFileFullPath)
					isExit = false
				}
			case StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					server.SetKeepAlivesEnabled(false)
					server.Shutdown(context.Background())
					err = deleteControlChannel(listenAddr.String())
					if nil != err {
						common.Error("Delete control channel fial, erro: %v", err)
					}
					isExit = true
				}
			default:
				{
					isExit = false
				}
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}(serverName, listenAddr)
}

// support HTTP1.1 and HTTP2.0
func initHttp2Server(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func(serverName string, listenAddr common.IpAndPort) {
		common.Debug("%v server control coroutine running...\n", serverName)
		server := newServer(listenAddr)
		err := http2.ConfigureServer(&server, &http2.Server{})
		if nil != err {
			panic(err)
		}

		c := make(chan int)
		err = insertControlChannel(listenAddr.String(), c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case StartServerControlOption:
				{
					common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
					go server.ListenAndServeTLS(certificateFileFullPath, privateKeyFileFullPath)
					isExit = false
				}
			case StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					server.SetKeepAlivesEnabled(false)
					server.Shutdown(context.Background())
					err = deleteControlChannel(listenAddr.String())
					if nil != err {
						common.Error("Delete control channel fial, erro: %v", err)
					}
					isExit = true
				}
			default:
				{
					isExit = false
				}
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}(serverName, listenAddr)
}

// HTTP version 3 over QUIC
func initHttp3Server(serverName string, listenAddr common.IpAndPort) {
	// control coroutine
	go func(serverName string, listenAddr common.IpAndPort) {
		common.Debug("%v server control coroutine running...\n", serverName)
		server := newServer(listenAddr)
		http3server := http3.Server{
			Addr:       listenAddr.Ip,
			Port:       int(listenAddr.Port),
			QuicConfig: &quic.Config{},
		}

		c := make(chan int)
		err := insertControlChannel(listenAddr.String(), c)
		if nil != err {
			panic(err)
		}

		isExit := false
		for {
			option := <-c
			switch option {
			case StartServerControlOption:
				{
					common.System("%v server startup, listen on %v\n", serverName, listenAddr.String())
					go http3server.ListenAndServeTLS(certificateFileFullPath, privateKeyFileFullPath)
					isExit = false
				}
			case StopServerControlOption:
				{
					common.System("%v server stop\n", serverName)
					server.SetKeepAlivesEnabled(false)
					server.Shutdown(context.Background())
					err = deleteControlChannel(listenAddr.String())
					if nil != err {
						common.Error("Delete control channel fial, erro: %v", err)
					}
					isExit = true
				}
			default:
				{
					isExit = false
				}
			}

			if isExit {
				break
			}
		}

		runtime.Goexit()
	}(serverName, listenAddr)
}

func newServer(listenAddr common.IpAndPort) http.Server {
	mux := http.NewServeMux()

	for _, filePrefix := range common.JsonConfigs.ServerDownloadPrefix {
		mux.Handle("/"+filePrefix+"/", http.StripPrefix("/"+filePrefix+"/", http.FileServer(http.Dir(common.CurrDir+"/"+filePrefix+"/"))))
	}

	mux.HandleFunc("/", index)
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/list", list)

	return http.Server{
		Addr:    listenAddr.String(),
		Handler: mux,
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	// receive
	recvBuffer := r.Header.Get("ClientSendData")
	remoteAddr := r.RemoteAddr

	// send
	w.Write([]byte(common.JsonConfigs.ServerSendData))
	w.Write([]byte("\n"))

	var prefix string
	if nil == r.TLS {
		prefix = "Http"
	} else {
		prefix = "Https"
	}
	if "Http" == prefix {
		serverHttpCount++
	} else {
		serverHttpsCount++
	}

	common.Info("%v server[%v]----%v client[%v]:\n\trecv: %s\n\tsend: %s\n", prefix, r.Host, prefix, remoteAddr, recvBuffer, common.JsonConfigs.ServerSendData)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.Create(uploadPath + handler.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	io.Copy(f, file)
	w.Write([]byte("Get upload page ok:\n\t"))
	fmt.Fprintf(w, "File [%v] upload to [%v] ok", handler.Filename, uploadPath+handler.Filename)
	common.Info("File [%v] upload to [%v] ok\n", handler.Filename, uploadPath+handler.Filename)
}

func list(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get list page ok:\n"))

	files, err := listDir(uploadPath, "")
	if nil != err {
		w.Write([]byte(fmt.Sprintf("\tList downloadable file names fail, %v", err.Error())))
		common.Warn("List downloadable file names fail, %v\n", err.Error())
		return
	}

	var listFile string
	for _, f := range files {
		line := fmt.Sprintf("\t\t%v\n", f)
		listFile += line
	}

	if 0 == len(listFile) {
		w.Write([]byte("\t\tNot have any file on web server"))
	} else {
		w.Write([]byte(fmt.Sprintf("\tTotal %v files, see below:\n", len(files))))
		w.Write([]byte(listFile))
	}

	common.Info("Get list page ok\n")
}

// 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func listDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	//忽略后缀匹配的大小写
	suffix = strings.ToUpper(suffix)
	for _, fi := range dir {
		// 忽略目录
		if fi.IsDir() {
			continue
		}

		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func HttpServerGuide(listenPort uint16) {
	ip := "127.0.0.1"
	common.System("Http server use guide:\n")
	common.System("\tUse 'curl http://%v:%v' get index page\n", ip, listenPort)
	common.System("\tUse 'curl -F \"uploadfile=@/filepath/filename\" http://%v:%v/upload' upload file to web server\n", ip, listenPort)
	common.System("\tUse 'curl http://%v:%v/list' list downloadable file names\n", ip, listenPort)
	common.System("\tUse 'wget http://%v:%v/files/filename' download file\n", ip, listenPort)
}

func HttpsServerGuide(listenPort uint16) {
	ip := "127.0.0.1"
	common.System("Https server certificate has been generated in the %v directory\n", certificatePath)
	common.System("Https server use guide:\n")
	common.System("\tUse 'curl -k https://%v:%v' get index page\n", ip, listenPort)
	common.System("\tUse 'curl -k --http1.1 https://%v:%v' get index page using HTTP1.1\n", ip, listenPort)
	common.System("\tUse 'curl -k --http2 https://%v:%v' get index page using HTTP2\n", ip, listenPort)
	common.System("\tUse 'curl -k --http3 https://%v:%v' get index page using HTTP3\n", ip, listenPort)
	common.System("\tUse 'curl -k -F \"uploadfile=@/filepath/filename\" https://%v:%v/upload' upload file to web server\n", ip, listenPort)
	common.System("\tUse 'curl -k https://%v:%v/list' list downloadable file names\n", ip, listenPort)
	common.System("\tUse 'wget --no-check-certificate https://%v:%v/files/filename' download file\n", ip, listenPort)
}

func preparePrivateKeyAndCert() {
	err := common.Rm(privateKeyFileFullPath)
	if nil != err {
		panic(err)
	}

	err = common.Rm(certificateFileFullPath)
	if nil != err {
		panic(err)
	}

	// Two steps generation privary key and crt:
	// 	openssl genrsa -out %v 2048 > /dev/null
	//		openssl req -new -x509 -key %v -out %v -days 365 -subj /C=CN/ST=Some-State/O=Internet > /dev/null

	cmd := fmt.Sprintf(`openssl req -newkey rsa:2048 -nodes -keyout %v -x509 -out %v -days 365 -subj /C=CN/ST=Some-State/O=Internet > /dev/null`, privateKeyFileFullPath, certificateFileFullPath)
	_, err = common.Command("/bin/sh", "-c", cmd)
	if nil != err {
		if !common.IsExist(privateKeyFileFullPath) || !common.IsExist(certificateFileFullPath) {
			panic(fmt.Sprintf("The file %v or %v not exist", privateKeyFileFullPath, certificateFileFullPath))
		}
	}
}
