#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import socket
import threading
import time
import subprocess
import urllib.request

from optparse import OptionParser
from optparse import OptionGroup
from http.server import HTTPServer, BaseHTTPRequestHandler

# Client configuration as below
ServerListenHost = "0.0.0.0"
ServerTcpListenPort1 = 1010
ServerUdpListenPort1 = 1020
ServerQuicListenPort1 = 1030
ServerHttpListenPort1 = 1080
ServerHttpsListenPort1 = 1443
ServerTcpListenPort2 = 2010
ServerUdpListenPort2 = 2020
ServerQuicListenPort2 = 2030
ServerHttpListenPort2 = 2080
ServerHttpsListenPort2 = 2443
ServerSendData = "Hello Client"

# Client configuration as below
ClientBindIpAddress = "127.0.0.1"
ClientSendToIpv4Address = "127.0.0.1"
ClientSendToIpv6Address = "::1"
ClientSendData = "Hello Server"

RecvBufferSizeBytes = 512
VERSION = "v1.0"
__author__ = "Wei Jian"


def parseOption():
    parser = OptionParser("Usage: %prog [options] arg1 arg2 ...")
    parser.add_option("-v", "--version", action="store_true", dest="isShowVersion", default=False, help="show version")

    serverGroup = OptionGroup(parser, 'Server Options', 'support for the following server options.')
    serverGroup.add_option("-s", "--server", action="store_true", dest="isRunAsServer", default=False,
                           help="the tester run as server, default run as client")
    parser.add_option_group(serverGroup)

    clientGroup = OptionGroup(parser, 'Client Options', 'support for the following client options.')
    clientGroup.add_option("-b", "--bind", action="store", type="string", dest="bindClientIpAddress", default="",
                           help="the IP address of client bind, this option takes precedence over 'ClientBindIpAddress'")
    clientGroup.add_option("-w", "--wait", action="store", type="int", dest="waitSecondBeforeSend", default=0,
                           help="the second waiting to send before")
    clientGroup.add_option("-n", "--number", action="store", type="int", dest="sendNumber", default=0,
                           help="the number of client send data to server, valid only for UDP, TCP, QUIC protocols")
    clientGroup.add_option("-d", "--dport", action="store", type="int", dest="destinationPort", default=0,
                           help="the destination port of client")
    clientGroup.add_option("--tcp", action="store_true", dest="isUsingTcp", default=False,
                           help="using TCP protocol as client")
    clientGroup.add_option("--udp", action="store_true", dest="isUsingUdp", default=False,
                           help="using Udp protocol as client")
    clientGroup.add_option("--quic", action="store_true", dest="isUsingQuic", default=False,
                           help="using QUIC protocol as client")
    clientGroup.add_option("--http", action="store_true", dest="isUsingHttp", default=False,
                           help="using HTTP protocol as client")
    clientGroup.add_option("--https", action="store_true", dest="isUsingHttps", default=False,
                           help="using HTTPs protocol as client")

    parser.add_option_group(clientGroup)
    (options, args) = parser.parse_args()

    if options.isRunAsServer:
        return options, args

    if 0 != options.destinationPort:
        if False == options.isUsingTcp and False == options.isUsingUdp and False == options.isUsingQuic and False == options.isUsingHttp and False == options.isUsingHttps:
            if options.destinationPort == ServerTcpListenPort1 or options.destinationPort == ServerTcpListenPort2:
                options.isUsingTcp = True
            elif options.destinationPort == ServerUdpListenPort1 or options.destinationPort == ServerUdpListenPort2:
                options.isUsingUdp = True
            elif options.destinationPort == ServerQuicListenPort1 or options.destinationPort == ServerQuicListenPort2:
                options.isUsingQuic = True
            elif options.destinationPort == ServerHttpListenPort1 or options.destinationPort == ServerHttpListenPort2:
                options.isUsingHttp = True
            elif options.destinationPort == ServerHttpsListenPort1 or options.destinationPort == ServerHttpsListenPort2:
                options.isUsingHttps = True
            else:
                print(
                    "The destination port has been specified, but you also need to specify the protocol used by the client.")
                exit(-1)

    elif options.isUsingTcp:
        options.destinationPort = ServerTcpListenPort1
    elif options.isUsingUdp:
        options.destinationPort = ServerUdpListenPort1
    elif options.isUsingQuic:
        options.destinationPort = ServerQuicListenPort1
    elif options.isUsingHttp:
        options.destinationPort = ServerHttpListenPort1
    elif options.isUsingHttps:
        options.destinationPort = ServerHttpsListenPort1
    else:
        print("Please specify the protocol or destination port to use.")
        exit(-1)

    if "" != options.bindClientIpAddress:
        if not isIpv4(options.bindClientIpAddress) and not isIpv6(options.bindClientIpAddress):
            print("Please specify the IP address of client binding.")
            exit(-1)
    else:
        options.bindClientIpAddress = ClientBindIpAddress

    return options, args


def isIpv4(ip):
    try:
        socket.inet_pton(socket.AF_INET, ip)
    except AttributeError:
        try:
            socket.inet_aton(ip)
        except socket.error:
            return False
        return ip.count('.') == 3
    except socket.error:
        return False
    return True


def isIpv6(ip):
    try:
        socket.inet_pton(socket.AF_INET6, ip)
    except socket.error:
        return False
    return True


class Server:
    serverName = ""
    listenIp = ""
    listenPort = 0
    _sock = -1

    def __init__(self, name, ip, port):
        self.serverName = name
        self.listenIp = ip
        self.listenPort = port

    def __del__(self):
        if -1 != self._sock:
            self._sock.close()

    def run(self):
        pass


class TcpServer(Server, threading.Thread):
    handleFunc = 0

    def __init__(self, name, ip, port, func=0):
        Server.__init__(self, name, ip, port)
        threading.Thread.__init__(self)
        self.handleFunc = func

    def run(self):
        _sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        _sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        _sock.bind((self.listenIp, self.listenPort))
        _sock.listen(5)
        print("Tcp server[{0}] startup, listen on {1}:{2}".format(self.serverName, self.listenIp, self.listenPort))
        while True:
            conn, addr = _sock.accept()
            func = self.handleFunc
            if 0 == self.handleFunc:
                func = self.handle
            try:
                t = threading.Thread(target=func, args=(conn,))
                t.start()
            except socket.error:
                pass

    def handle(self, conn):
        time = 0
        while True:
            data = conn.recv(RecvBufferSizeBytes)
            conn.send(ServerSendData.encode())
            ++times
            print("Tcp server {0} [{1}]---------Tcp client[{2}], times[{3}]:\n\trecv: {4}\n\tsend: {5}\n".format(
                self.serverName,
                conn.getsockname(),
                conn.getpeername(),
                times,
                data.decode(),
                ServerSendData))
        conn.close()


class UdpServer(Server, threading.Thread):
    def __init__(self, name, ip, port):
        Server.__init__(self, name, ip, port)
        threading.Thread.__init__(self)

    def run(self):
        _sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        _sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        _sock.bind((self.listenIp, self.listenPort))
        print("Udp server[{0}] startup, listen on {1}:{2}".format(self.serverName, self.listenIp, self.listenPort))

        times = 0
        while True:
            data, addr = _sock.recvfrom(RecvBufferSizeBytes)
            _sock.sendto(ServerSendData.encode(), addr)
            ++times
            print("Udp server {0} [{1}]---------Udp client[{2}], times[{3}]:\n\trecv: {4}\n\tsend: {5}\n".format(
                self.serverName,
                _sock.getsockname(),
                addr,
                times,
                data.decode(),
                ServerSendData))


class Resquest(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/text')
        self.end_headers()
        self.wfile.write(ServerSendData.encode())


class HttpServer(Server):
    httpServer = 0

    def __init__(self, name, ip, port):
        Server.__init__(self, name, ip, port)
        self.httpServer = HTTPServer((ip, port,), Resquest)

    def start(self):
        print("Http server[{0}] startup, listen on {1}:{2}".format(self.serverName, self.listenIp, self.listenPort))
        self.httpServer.server_activate()


class HttpsServer(Server):
    keyFullPath = "/tmp/tester/tester.key"
    crtFullPath = "/tmp/tester/tester.crt"
    httpsServer = 0

    def __init__(self, name, ip, port):
        Server.__init__(self, name, ip, port)
        self._genCert()
        self.httpsServer = HTTPServer((ip, port,), Resquest)
        self.httpsServer.socket = ssl.wrap_socket(httpd.socket, keyfile=self.keyFullPath, certfile=self.crtFullPath,
                                                  server_side=True)

    def start(self):
        print("Https server[{0}] startup, listen on {1}:{2}".format(self.serverName, self.listenIp, self.listenPort))
        self.httpsServer.server_activate()

    def _genCert(self):
        if not os.path.exists(self.keyFullPath):
            cmd1 = "openssl genrsa -out {0} 2048 > /dev/null".format(self.keyFullPath)
            cmd2 = "openssl req -new -x509 -key {0} -out {1} -days 365 -subj /C=CN/ST=Some-State/O=Internet > /dev/null".format(
                self.keyFullPath, self.crtFullPath)
            ret = subprocess.run(cmd1)
            if 0 != ret.returncode:
                print("Generate Https certificate failed1")
                exit(-1)
            ret = subprocess.run(cmd2)
            if 0 != ret.returncode:
                print("Generate Https certificate failed2")
                exit(-1)


class Client:
    bindIp = ""
    destAddr = ()
    sendCounter = 1
    waitSecondBeforeSend = 0
    _sock = -1

    def __init__(self, srcIp, destIp, destPort, n=1, wait=0):
        self.bindIp = srcIp
        self.destAddr = (destIp, destPort)
        self.sendCounter = n
        self.waitSecondBeforeSend = wait

    def __del__(self):
        if -1 != self._sock:
            self._sock.close()
            self._sock = -1

    def start(self):
        pass


class TcpClient(Client):
    def __init__(self, srcIp, destIp, destPort, n=1, wait=0):
        Client.__init__(self, srcIp, destIp, destPort, n, wait)

    def start(self):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self._sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._sock.bind((self.bindIp, 0))
        self._sock.connect(self.destAddr)
        time.sleep(self.waitSecondBeforeSend)
        times = 0

        while 0 != self.sendCounter:
            self._sock.send(ClientSendData.encode())
            data = self._sock.recv(RecvBufferSizeBytes).decode()
            --self.sendCounter
            ++times
            print("Tcp client[{0}]---------Tcp server[{1}], times[{2}]:\n\tsend: {3}\n\trecv: {4}\n".format(
                self._sock.getsockname(), self._sock.getpeername(), times, ServerSendData, data))


class UdpClient(Client):
    def __init__(self, srcIp, destIp, destPort, n=1, wait=0):
        Client.__init__(self, srcIp, destIp, destPort, n, wait)

    def start(self):
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self._sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._sock.bind((self.bindIp, 0))
        self._sock.connect(self.destAddr)
        time.sleep(self.waitSecondBeforeSend)

        times = 0
        while 0 != self.sendCounter:
            self._sock.send(ClientSendData.encode())
            data = self._sock.recv(RecvBufferSizeBytes).decode()
            --self.sendCounter
            ++times
            print("Udp client[{0}]---------Udp server[{1}], times[{2}]:\n\tsend: {3}\n\trecv: {4}\n".format(
                self._sock.getsockname(), self._sock.getpeername(), times, ServerSendData, data))


class HttpClient(Client):
    url = ""

    def __init__(self, srcIp, destIp, destPort, n=1, wait=0):
        Client.__init__(self, srcIp, destIp, destPort, n, wait)
        if 4 == isIpv4(self.bindIp):
            self.url = "http://{0}/".format(self.destAddr)
        elif 6 == isIp6(self.bindIp):
            self.url = "http://[{0}]/".format(self.destAddr)

    def start(self):
        times = 0
        while 0 != self.sendCounter:
            request = urllib.request.Request(url, data=ClientSendData.encode())
            response = urllib.request.urlopen(request)
            ++times
            data = response.read().decode('utf-8')
            print("Http client[{0}]---------Udp server[{1}], times[{2}]:\n\tsend: {3}\n\trecv: {4}\n".format(
                self.bindIp, self.destAddr, times, ServerSendData, data))


if __name__ == "__main__":
    (options, args) = parseOption()

    if options.isShowVersion:
        print("{0} version: {1}".format(os.path.basename(__file__), VERSION))
        exit(0)

    if options.isRunAsServer:
        tcpServer1 = TcpServer("tcp", ServerListenHost, ServerTcpListenPort1)
        tcpServer1.start()
        time.sleep(1)

        udpServer1 = UdpServer("udp", ServerListenHost, ServerUdpListenPort1)
        udpServer1.start()

        httpServer1 = HttpServer("http", ServerListenHost, ServerHttpListenPort1)
        httpServer1.start()

        # httpsServer1 = HttpsServer("https", ServerListenHost, ServerHttpsListenPort1)
        # httpsServer1.start()

        while True:
            time.sleep(10)

    elif options.isUsingTcp:
        tcpClient = 0
        if isIpv4(options.bindClientIpAddress):
            tcpClient = TcpClient(options.bindClientIpAddress, ClientSendToIpv4Address, options.destinationPort,
                                  options.sendNumber, options.waitSecondBeforeSend)
        else:
            tcpClient = TcpClient(options.bindClientIpAddress, ClientSendToIpv6Address, options.destinationPort,
                                  options.sendNumber, options.waitSecondBeforeSend)

        tcpClient.start()

    elif options.isUsingUdp:
        udpClient = 0
        if isIpv4(options.bindClientIpAddress):
            udpClient = UdpClient(options.bindClientIpAddress, ClientSendToIpv4Address, options.destinationPort,
                                  options.sendNumber, options.waitSecondBeforeSend)
        else:
            udpClient = UdpClient(options.bindClientIpAddress, ClientSendToIpv6Address, options.destinationPort,
                                  options.sendNumber, options.waitSecondBeforeSend)

        udpClient.start()

    elif options.isUsingQuic:
        pass
    elif options.isUsingHttp:
        pass
    elif options.isUsingHttps:
        pass
    else:
        pass
