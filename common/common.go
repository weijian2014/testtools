package common

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func IsLocalIP(ip string) (bool, error) {
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

func Command(name string, arg ...string) ([]byte, error) {
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
		return []byte(""), fmt.Errorf(noEnterErrStr)
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

//RSA公钥私钥产生 无法启动HTTPS服务器
func GenerateX509Certificate(keyFullPath string, crtFullPath string) error {
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

func GetIpAndPort(host string) (string, uint16, error) {
	if !strings.Contains(host, ":") {
		return "", 0, fmt.Errorf("host[%v] error", host)
	}

	index := strings.LastIndex(host, ":")
	ip := host[0:index]

	p, err := strconv.ParseUint(host[index+1:], 10, 16)
	if nil != err {
		return "", 0, fmt.Errorf("host[%v] parse fail, err: %v", host, err)
	}

	port := uint16(p)
	if 0 >= port || 65535 < port {
		return "", 0, fmt.Errorf("port of host[%v] invalid", host)
	}

	return ip, port, nil
}
