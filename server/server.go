package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go"
	"golang.org/x/net/dns/dnsmessage"
)

var (
	CurrDir                    = ""
	UploadPath                 = ""
	CertificatePath            = ""
	IsHelp                     = false
	Configs            *Config = nil
	ConfigFileFullPath         = ""
	DnsAEntrys         map[string]string
	Dns4AEntrys        map[string]string
)

type Config struct {
	// Common Config
	CommonRecvBufferSizeBytes uint64 `json:"CommonRecvBufferSizeBytes"`
	// Server Config
	ServerListenHost         string `json:"ServerListenHost"`
	ServerTcpListenPort      uint16 `json:"ServerTcpListenPort"`
	ServerUdpListenPort      uint16 `json:"ServerUdpListenPort"`
	ServerHttpListenPort     uint16 `json:"ServerHttpListenPort"`
	ServerHttpsListenPort    uint16 `json:"ServerHttpsListenPort"`
	ServerIeeeQuicListenPort uint16 `json:"ServerIeeeQuicListenPort"`
	ServerDnsListenPort      uint16 `json:"ServerDnsListenPort"`
	// map[string]interface{}
	ServerDnsAEntrys  interface{} `json:"ServerDnsAEntrys"`
	ServerDns4AEntrys interface{} `json:"ServerDns4AEntrys"`
	ServerSendData    string      `json:"ServerSendData"`
	// Client Config
	ClientBindIpAddress     string `json:"ClientBindIpAddress"`
	ClientSendToIpv4Address string `json:"ClientSendToIpv4Address"`
	ClientSendToIpv6Address string `json:"ClientSendToIpv6Address"`
	ClientSendData          string `json:"ClientSendData"`
}

func init() {
	currFullPath, err := exec.LookPath(os.Args[0])
	if nil != err {
		panic(err)
	}

	absFullPath, err := filepath.Abs(currFullPath)
	if nil != err {
		panic(err)
	}
	CurrDir = filepath.Dir(absFullPath)
	UploadPath = CurrDir + "/files/"
	CertificatePath = CurrDir + "/cert/"

	flag.BoolVar(&IsHelp, "h", false, "Show help")
	flag.StringVar(&ConfigFileFullPath, "f", CurrDir+"/config.json", "The path of config.json file, support for absolute and relative paths")
	flag.Parse()
}

func main() {
	var err error
	_, err = os.Stat(ConfigFileFullPath)
	if os.IsNotExist(err) {
		ConfigFileFullPath = CurrDir + "/../config/config.json"
	}
	Configs, err = loadConfigFile(ConfigFileFullPath)
	if nil != err {
		panic(err)
	}
	err = checkConfigFlie()
	if nil != err {
		panic(err)
	}

	if IsHelp {
		flag.Usage()
		printHttpServerGuide()
		printHttpsServerGuide()
		saveDnsEntrys()
		printDnsServerEntrys()
		fmt.Printf("Json config: %+v\n\n", Configs)
		return
	}

	// 创建./files/目录
	_, err = os.Stat(UploadPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(UploadPath, os.ModePerm)
		if nil != err {
			panic(err)
		}
	}

	// 在./files/目录下创建一个test.txt文件， 并写入ServerSendData数据
	testFile, err := os.Create(UploadPath + "test.txt")
	if nil != err {
		panic(err)
	}
	testFile.Write([]byte(Configs.ServerSendData))
	testFile.Write([]byte("\n"))
	testFile.Close()

	go startTcpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startUdpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startHttpsServer()
	time.Sleep(time.Duration(200) * time.Millisecond)

	go startDnsServer()
	time.Sleep(time.Duration(50) * time.Millisecond)

	go startIeeeQuicServer()

	time.Sleep(time.Duration(200) * time.Millisecond)
	printHttpServerGuide()
	printHttpsServerGuide()
	printDnsServerEntrys()

	for {
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func startTcpServer() {
	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerTcpListenPort)
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tcp   server startup, listen on %v\n", serverAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Tpc server accept failed, err: %v\n", err)
			continue
		}

		go newTcpConnectionHandler(conn)
	}
}

func startUdpServer() {
	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerUdpListenPort)
	udp, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Udp   server startup, listen on %v\n", serverAddress)

	for {
		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			fmt.Printf("Udp server[%v]----Udp client[%v] receive failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		// send
		n, err := conn.WriteToUDP([]byte(Configs.ServerSendData), remoteAddress)
		if err != nil {
			fmt.Printf("Udp server[%v]----Udp client[%v] send failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		fmt.Printf("Udp server[%v]----Udp client[%v]:\n\trecv: %s\n\tsend: %s\n", conn.LocalAddr(), remoteAddress, recvBuffer[:n], Configs.ServerSendData)
	}
}

func startHttpServer() {
	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerHttpListenPort)
	fmt.Printf("Http  server startup, listen on %v\n", serverAddress)

	// 启动静态文件服务, 将下载服务器存放文件的目录
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(UploadPath))))

	// 页面
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/list", list)
	http.ListenAndServe(serverAddress, nil)
}

func startHttpsServer() {
	_, err := os.Stat(CertificatePath)
	if os.IsNotExist(err) {
		err = os.Mkdir(CertificatePath, os.ModePerm)
		if nil != err {
			panic(err)
		}
	}

	keyFullPath := CertificatePath + "server.key"
	crtFullPath := CertificatePath + "server.crt"

	_, err = os.Stat(keyFullPath)
	if os.IsExist(err) {
		err = os.Remove(keyFullPath)
		if nil != err {
			panic(err)
		}
	}

	_, err = os.Stat(crtFullPath)
	if os.IsExist(err) {
		err = os.Remove(crtFullPath)
		if nil != err {
			panic(err)
		}
	}
	err = generateHttpsCertificate(keyFullPath, crtFullPath)
	if nil != err {
		panic(err)
	}

	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerHttpsListenPort)
	fmt.Printf("Https server startup, listen on %v\n", serverAddress)

	// 启动静态文件服务 Http已经注册过
	//http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(UploadPath))))

	// 页面 Http已经注册过
	//http.HandleFunc("/", index)
	//http.HandleFunc("/upload", upload)
	//http.HandleFunc("/list", list)
	http.ListenAndServeTLS(serverAddress, crtFullPath, keyFullPath, nil)
}

func startDnsServer() {
	saveDnsEntrys()

	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerDnsListenPort)
	udp, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", udp)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Dns   server startup, listen on %v\n", serverAddress)

	for {
		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		_, remoteAddress, err := conn.ReadFromUDP(recvBuffer)
		if err != nil {
			fmt.Printf("Dns server[%v]----Dns client[%v] receive failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		var requestMessage dnsmessage.Message
		err = requestMessage.Unpack(recvBuffer)
		if nil != err {
			fmt.Printf("Dns server[%v]----Dns client[%v] unpack failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		//fmt.Printf("Dns server[%v]----Dns client[%v], recv msg:\n\t%+v\n", conn.LocalAddr(), remoteAddress, requestMessage)
		questionCount := len(requestMessage.Questions)
		if 0 == questionCount {
			fmt.Printf("Dns server[%v]----Dns client[%v] question count is zero\n", conn.LocalAddr(), remoteAddress)
			continue
		} else {
			requestMessage.Header.Response = true
			requestMessage.Header.Authoritative = true
		}

		var answers []dnsmessage.Resource
		var tmp string
		for _, question := range requestMessage.Questions {
			h := dnsmessage.ResourceHeader{
				Name:  question.Name,
				Type:  question.Type,
				Class: question.Class,
				TTL:   3600,
			}

			if dnsmessage.TypeA == question.Type {
				ipv4, isOk := DnsAEntrys[question.Name.String()]
				if isOk {
					ip := net.ParseIP(ipv4).To4()
					b := &dnsmessage.AResource{
						A: [4]byte{ip[0], ip[1], ip[2], ip[3]},
					}
					answers = append(answers, dnsmessage.Resource{Header: h, Body: b})
					tmp += ipv4
					tmp += ", "
				}
			} else if dnsmessage.TypeAAAA == question.Type {
				ipv6, isOk := Dns4AEntrys[question.Name.String()]
				if isOk {
					ip := net.ParseIP(ipv6).To16()
					b := &dnsmessage.AAAAResource{
						AAAA: [16]byte{
							byte(ip[0]), byte(ip[1]), byte(ip[2]), byte(ip[3]),
							byte(ip[4]), byte(ip[5]), byte(ip[6]), byte(ip[7]),
							byte(ip[8]), byte(ip[9]), byte(ip[10]), byte(ip[11]),
							byte(ip[12]), byte(ip[13]), byte(ip[14]), byte(ip[15]),
						},
					}
					answers = append(answers, dnsmessage.Resource{Header: h, Body: b})
					tmp += ipv6
					tmp += ", "
				}
			} else {
				//fmt.Printf("Dns server[%v]----Dns client[%v] question[%d] is not A or AAAA\n", conn.LocalAddr(), remoteAddress, i+1)
				continue
			}

			requestMessage.Answers = answers
		}

		// send
		packed, err := requestMessage.Pack()
		if nil != err {
			fmt.Printf("Dns server[%v]----Dns client[%v] pack failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}
		_, err = conn.WriteToUDP(packed, remoteAddress)
		if err != nil {
			fmt.Printf("Dns server[%v]----Dns client[%v] send failed, err : %v\n", conn.LocalAddr(), remoteAddress, err)
			continue
		}

		tmp = strings.TrimRight(tmp, ", ")
		fmt.Printf("Dns server[%v]----Dns client[%v]:\n\tquestion: %+v\n\t answers: %+v\n",
			conn.LocalAddr(), remoteAddress, requestMessage.Questions, tmp)
	}
}

func startIeeeQuicServer() {
	serverAddress := fmt.Sprintf("%v:%v", Configs.ServerListenHost, Configs.ServerIeeeQuicListenPort)
	listener, err := quic.ListenAddr(serverAddress, generateQuicTLSConfig(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Quic  server startup, listen on %v\n", serverAddress)

	for {
		session, err := listener.Accept()
		if err != nil {
			fmt.Printf("Quic server accept fail, err: %v\n", err)
			continue
		}

		go newQuicSessionHandler(session)
	}
}

func newTcpConnectionHandler(conn net.Conn) {
	defer conn.Close()
	for {
		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		n, err := conn.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() {
				break
			}

			if "EOF" == err.Error() {
				break
			}

			fmt.Printf("Tcp server[%v]----Tcp client[%v] receive failed, err: %v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		// send
		_, err = conn.Write([]byte(Configs.ServerSendData))
		if nil != err {
			fmt.Printf("Tcp server[%v]----Tcp client[%v] send failed, err: %v\n", conn.LocalAddr(), conn.RemoteAddr(), err)
			break
		}

		fmt.Printf("Tcp server[%v]----Tcp client[%v]:\n\trecv: %s\n\tsend: %s\n", conn.LocalAddr(), conn.RemoteAddr(), recvBuffer[:n], Configs.ServerSendData)
	}

	fmt.Printf("Tcp server[%v]----Tcp client[%v] closed\n", conn.LocalAddr(), conn.RemoteAddr())
}

func newQuicSessionHandler(sess quic.Session) {
	stream, err := sess.AcceptStream()
	defer stream.Close()
	if err != nil {
		fmt.Printf("Quic server[%v] ---- %v accept stream failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
		return
	}

	for {
		// receive
		recvBuffer := make([]byte, Configs.CommonRecvBufferSizeBytes)
		_, err = stream.Read(recvBuffer)
		if err != nil {
			if "NO_ERROR" == err.Error() {
				break
			}

			if "EOF" == err.Error() {
				break
			}

			fmt.Printf("Quic server[%v]----Quic client[%v] receive failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		// send
		n, err := stream.Write([]byte(Configs.ServerSendData))
		if nil != err {
			fmt.Printf("Quic server[%v]----Quic client[%v] send failed, err: %v\n", sess.LocalAddr(), sess.RemoteAddr(), err)
			return
		}

		fmt.Printf("Quic server[%v]----Quic client[%v]:\n\trecv: %s\n\tsend: %s\n", sess.LocalAddr(), sess.RemoteAddr(), recvBuffer[:n], Configs.ServerSendData)
	}

	fmt.Printf("Quic server[%v]----Quic client[%v] closed\n", sess.LocalAddr(), sess.RemoteAddr())
}

// Setup a bare-bones TLS config for the server
func generateQuicTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	// receive
	recvBuffer := r.Header.Get("ClientSendData")
	remoteAddr := r.RemoteAddr
	var prefix string
	if nil == r.TLS {
		prefix = "Http"
	} else {
		prefix = "Https"
	}

	// send
	w.Write([]byte(Configs.ServerSendData))
	w.Write([]byte("\n"))
	fmt.Printf("%v server[%v]----%v client[%v]:\n\trecv: %s\n\tsend: %s\n", prefix, r.Host, prefix, remoteAddr, recvBuffer, Configs.ServerSendData)
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	f, err := os.Create(UploadPath + handler.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	io.Copy(f, file)
	w.Write([]byte("Get upload page ok:\n\t"))
	fmt.Fprintf(w, "File [%v] upload to [%v] ok", handler.Filename, UploadPath+handler.Filename)
	fmt.Printf("File [%v] upload to [%v] ok\n", handler.Filename, UploadPath+handler.Filename)
}

func list(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get list page ok:\n"))

	files, err := listDir(UploadPath, "")
	if nil != err {
		w.Write([]byte(fmt.Sprintf("\tList downloadable file names fail, %v", err.Error())))
		fmt.Printf("List downloadable file names fail, %v\n", err.Error())
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

	fmt.Printf("Get list page ok\n")
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

func generateHttpsCertificate(keyFullPath string, crtFullPath string) error {
	cmd := fmt.Sprintf("openssl genrsa -out %v 2048 > /dev/null", keyFullPath)
	_, err := command("/bin/sh", "-c", cmd)
	if nil != err {
		_, err1 := os.Stat(keyFullPath)
		if os.IsNotExist(err1) {
			return err
		}
	}

	cmd = fmt.Sprintf(`openssl req -new -x509 -key %v -out %v -days 365 -subj /C=CN/ST=Some-State/O=Internet > /dev/null`, keyFullPath, crtFullPath)
	_, err = command("/bin/sh", "-c", cmd)
	if nil != err {
		return err
	}

	_, err = os.Stat(crtFullPath)
	if os.IsNotExist(err) {
		return err
	}

	return nil
}

//RSA公钥私钥产生 无法启动HTTPS服务器
func generateX509Certificate(keyFullPath string, crtFullPath string) error {
	max := new(big.Int).Lsh(big.NewInt(1), 128)   //把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max) //返回在 [0, max) 区间均匀随机分布的一个随机值

	subject := pkix.Name{ //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
		Organization:       []string{"Https Publications Co."},
		OrganizationalUnit: []string{"Books"},
		CommonName:         "Https CN",
	}

	template := x509.Certificate{
		SerialNumber: serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, //KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},               // 密钥扩展用途的序列
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	pk, err := rsa.GenerateKey(rand.Reader, 2048) //生成一对具有指定字位数的RSA密钥
	if nil != err {
		panic(err)
	}

	//CreateCertificate基于模板创建一个新的证书, 第二个第三个参数相同，则证书是自签名的, 返回的切片是DER编码的证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	if nil != err {
		panic(err)
	}

	certOut, err := os.Create(crtFullPath)
	if nil != err {
		panic(err)
	}

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICAET", Bytes: derBytes})
	if nil != err {
		panic(err)
	}
	certOut.Close()

	keyOut, err := os.Create(keyFullPath)
	if nil != err {
		panic(err)
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	if nil != err {
		panic(err)
	}
	keyOut.Close()

	return nil
}

func command(name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, arg...)

	// 使用CommandContext而不用Command,是因为Command执行后无法kill掉shell进程
	// cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		cancel()
		cmd.Wait()
		return []byte(""), err
	}

	stderr, err := cmd.StderrPipe()
	defer stderr.Close()
	if err != nil {
		cancel()
		cmd.Wait()
		return []byte(""), err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		cmd.Wait()
		return []byte(""), err
	}

	errBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		cancel()
		cmd.Wait()
		return []byte(""), err
	}

	if 0 != len(errBytes) {
		noEnterErrStr := strings.TrimRight(string(errBytes), "\n")
		cancel()
		cmd.Wait()
		return []byte(""), errors.New(noEnterErrStr)
	}

	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		cancel()
		cmd.Wait()
		return []byte(""), err
	}

	cancel()
	cmd.Wait()

	return opBytes[:], nil
}

// 读取json配置文件
func loadConfigFile(filePath string) (*Config, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := json.Unmarshal(cData, c); nil != err {
		return nil, err
	}

	return c, nil
}

func checkConfigFlie() error {
	if "" != Configs.ServerListenHost &&
		"localhost" != Configs.ServerListenHost &&
		"0.0.0.0" != Configs.ServerListenHost &&
		"127.0.0.1" != Configs.ServerListenHost &&
		"::" != Configs.ServerListenHost {
		isLocal, err := isLocalIP(Configs.ServerListenHost)
		if nil != err {
			return err
		} else if !isLocal {
			return errors.New(fmt.Sprintf("ServerListenHost[%v] is not local address of config.json file", Configs.ServerListenHost))
		}
	}

	if 0 > Configs.ServerTcpListenPort || 65535 < Configs.ServerTcpListenPort {
		return errors.New(fmt.Sprintf("ServerTcpListenPort[%v] invalid of config.json file", Configs.ServerTcpListenPort))
	}

	if 0 > Configs.ServerUdpListenPort || 65535 < Configs.ServerUdpListenPort {
		return errors.New(fmt.Sprintf("ServerUdpListenPort[%v] invalid of config.json file", Configs.ServerUdpListenPort))
	}

	if 0 > Configs.ServerHttpListenPort || 65535 < Configs.ServerHttpListenPort {
		return errors.New(fmt.Sprintf("ServerHttpListenPort[%v] invalid of config.json file", Configs.ServerHttpListenPort))
	}

	if 0 > Configs.ServerHttpsListenPort || 65535 < Configs.ServerHttpsListenPort {
		return errors.New(fmt.Sprintf("ServerHttpsListenPort[%v] invalid of config.json file", Configs.ServerHttpsListenPort))
	}

	if 0 > Configs.ServerIeeeQuicListenPort || 65535 < Configs.ServerIeeeQuicListenPort {
		return errors.New(fmt.Sprintf("ServerIeeeQuicListenPort[%v] invalid of config.json file", Configs.ServerIeeeQuicListenPort))
	}

	if 0 > Configs.ServerDnsListenPort || 65535 < Configs.ServerDnsListenPort {
		return errors.New(fmt.Sprintf("ServerDnsListenPort[%v] invalid of config.json file", Configs.ServerDnsListenPort))
	}

	return nil
}

func printHttpServerGuide() {
	ip := net.ParseIP(Configs.ClientSendToIpv4Address)
	fmt.Printf("Http server use guide:\n")
	fmt.Printf("\tUse 'curl http://%v:%v' get index page\n", ip, Configs.ServerHttpListenPort)
	fmt.Printf("\tUse 'curl -F \"uploadfile=@/filepath/filename\" http://%v:%v/upload' upload file to web server\n", ip, Configs.ServerHttpListenPort)
	fmt.Printf("\tUse 'curl http://%v:%v/list' list downloadable file names\n", ip, Configs.ServerHttpListenPort)
	fmt.Printf("\tUse 'wget http://%v:%v/files/filename' download file\n", ip, Configs.ServerHttpListenPort)
}

func printHttpsServerGuide() {
	ip := net.ParseIP(Configs.ClientSendToIpv4Address)
	fmt.Printf("Https server certificate has been generated in the %v directory\n", CertificatePath)
	fmt.Printf("Https server use guide:\n")
	fmt.Printf("\tUse 'curl -k https://%v:%v' get index page\n", ip, Configs.ServerHttpsListenPort)
	fmt.Printf("\tUse 'curl -k -F \"uploadfile=@/filepath/filename\" https://%v:%v/upload' upload file to web server\n", ip, Configs.ServerHttpsListenPort)
	fmt.Printf("\tUse 'curl -k https://%v:%v/list' list downloadable file names\n", ip, Configs.ServerHttpsListenPort)
	fmt.Printf("\tUse 'wget --no-check-certificate https://%v:%v/files/filename' download file\n", ip, Configs.ServerHttpsListenPort)
}

func printDnsServerEntrys() {
	if 0 != len(DnsAEntrys) {
		fmt.Printf("Dns server a record:\n")
	}
	for k, v := range DnsAEntrys {
		fmt.Printf("\t%v ---- %v\n", k, v)
	}

	if 0 != len(Dns4AEntrys) {
		fmt.Printf("Dns server aaaa record:\n")
	}
	for k, v := range Dns4AEntrys {
		fmt.Printf("\t%v ---- %v\n", k, v)
	}
	fmt.Printf("\n")
}

func checkDomainName(domainName string) error {
	if strings.Contains(domainName, " ") {
		return errors.New(fmt.Sprintf("The domain name %v invalid", domainName))
	}

	if strings.HasPrefix(domainName, "http") {
		return errors.New(fmt.Sprintf("The domain name %v invalid, the prefix has 'http'", domainName))
	}

	if strings.HasPrefix(domainName, "https") {
		return errors.New(fmt.Sprintf("The domain name %v invalid, the prefix has 'https", domainName))
	}

	//支持以http://或者https://开头并且域名中间有/的情况
	isLine := "^((http://)|(https://))?([a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,6}(/)"
	_, err := regexp.MatchString(isLine, domainName)
	if nil != err {
		return err
	}

	//支持以http://或者https://开头并且域名中间没有/的情况
	notLine := "^((http://)|(https://))?([a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,6}"
	_, err = regexp.MatchString(notLine, domainName)
	if nil != err {
		return err
	}

	_, err = url.Parse(domainName)
	if nil != err {
		return err
	}

	return nil
}

func saveDnsEntrys() {
	// 读取配置文件中的A记录到map<domainName, IPv4>
	aEntryMap := Configs.ServerDnsAEntrys.(map[string]interface{})
	DnsAEntrys = make(map[string]string, len(aEntryMap)+1)
	for domainName, ip := range aEntryMap {
		if nil != checkDomainName(domainName) {
			panic(fmt.Sprintf("The domain name %v invalid", domainName))
		}

		ipv4 := ip.(string)
		if nil == net.ParseIP(ipv4) ||
			false == strings.Contains(ipv4, ".") {
			panic(fmt.Sprintf("The domain name %v not match valid IPv4 address", domainName))
		}
		DnsAEntrys[domainName+"."] = ipv4
	}
	DnsAEntrys["www.example.com."] = "127.0.0.1"

	// 读取配置文件中的AAAA记录到map<domainName, IPv6>
	aaaaEntryMap := Configs.ServerDns4AEntrys.(map[string]interface{})
	Dns4AEntrys = make(map[string]string, len(aaaaEntryMap)+1)
	for domainName, ip := range aaaaEntryMap {
		if nil != checkDomainName(domainName) {
			panic(fmt.Sprintf("The domain name %v invalid", domainName))
		}

		ipv6 := ip.(string)
		if nil == net.ParseIP(ipv6) ||
			false == strings.Contains(ipv6, ":") {
			panic(fmt.Sprintf("The domain name %v not match valid IPv6 address", domainName))
		}
		Dns4AEntrys[domainName+"."] = ipv6
	}
	Dns4AEntrys["www.example.com."] = "::1"
}

func isLocalIP(ip string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for i := range addrs {
		intf, _, err := net.ParseCIDR(addrs[i].String())
		if err != nil {
			return false, err
		}
		if net.ParseIP(ip).Equal(intf) {
			return true, nil
		}
	}
	return false, nil
}
