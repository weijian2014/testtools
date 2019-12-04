#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <getopt.h>
#include <cstdlib>
#include <iostream>
#include <string>

#define SIMULATOR_VERSION "v0.1"

using namespace std;

static const uint8_t IosSynPacketTcpOptions[24] = {
        0x02, 0x04, 0x05, 0xb4,
        0x01,
        0x03, 0x03, 0x07,
        0x01,
        0x01,
        0x08, 0x0a, 0x8a, 0x4f, 0xe5, 0xb0, 0x00, 0x00, 0x00, 0x00,
        0x04, 0x02,
        0x00,
        0x00,
};

//IP首部
struct IpHeader {
    uint8_t headerLength:4, ipVersion:4;    //4位首部长度+4位IP版本号
    uint8_t tos;                            //8位服务类型TOS
    uint16_t totalLength;                   //16位总长度（字节）
    uint16_t ident;                         //16位标识
    uint16_t off;                           //3位标志位
    uint8_t ttl;                            //8位生存时间 TTL
    uint8_t proto;                          //8位协议 (TCP, UDP 或其他)
    uint16_t checksum;                      //16位IP首部校验和
    uint32_t srcIp;                         //32位源IP地址
    uint32_t destIp;                        //32位目的IP地址
};

//IOS TCP首部
struct IosTcpHeader {
    uint16_t srcPort;                     //16位源端口
    uint16_t destPort;                    //16位目的端口
    uint32_t seq;                         //32位序列号
    uint32_t ack;                         //32位确认号
    uint8_t offset:4, headerLength:4;   //4位首部长度 6位保留字
    uint8_t flag;                         //6位标志位
    uint16_t winSize;                     //16位窗口大小
    uint16_t checksum;                    //16位校验和
    uint16_t surgent;                     //16位紧急数据偏移量
    uint8_t options[24];                  //IOS 24byte options
};

struct HelloTcpHeader {
    uint16_t srcPort;                     //16位源端口
    uint16_t destPort;                    //16位目的端口
    uint32_t seq;                         //32位序列号
    uint32_t ack;                         //32位确认号
    uint8_t offset:4, headerLength:4;   //4位首部长度 6位保留字
    uint8_t flag;                         //6位标志位
    uint16_t winSize;                     //16位窗口大小
    uint16_t checksum;                    //16位校验和
    uint16_t surgent;                     //16位紧急数据偏移量
};

//TCP伪首部
struct PsdHeader {
    uint32_t srcAddr;       //源地址
    uint32_t destAddr;      //目的地址
    uint8_t mbz;            //置空
    uint8_t proto;          //协议类型
    uint16_t tcpHeaderAndDataLength;     //TCP头+数据的长度
};

int enableRST(const string &srcIp, const string &destIp, uint16_t srcPort, uint16_t destPort) {
    char buf[256];
    sprintf(buf, "iptables -D OUTPUT -p tcp -s %s -d %s --sport %d --dport %d --tcp-flags ALL RST -j DROP",
            srcIp.c_str(), destIp.c_str(), srcPort, destPort);

    if (0 > ::system(buf)) {
        printf("Enable RST packet failed\n");
        return -1;
    } else {
        printf("Enable RST packet ok, cmd=%s\n", buf);
    }

    return 0;
}

int getAvaliablePort() {
    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = htonl(INADDR_ANY);
    addr.sin_port = 0;

    int sock = ::socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        printf("create socket failed\n");
        return -1;
    }

    if (0 != ::bind(sock, (sockaddr *) &addr, sizeof addr)) {
        printf("bind socket failed\n");
        return -1;
    }

    struct sockaddr_in sockAddr;
    int len = sizeof(sockAddr);
    if (0 != ::getsockname(sock, (sockaddr *) &sockAddr, (socklen_t *) &len)) {
        printf("getsockname socket failed\n");
        return -1;
    }

    uint16_t port(0);
    port = ntohs(sockAddr.sin_port);

    if (0 != ::close(sock)) {
        printf("close socket failed\n");
        return -1;
    }

    return port;
}

static const string HelloDataC("Hello Server");
static const size_t HelloDataLengthC(12);
static const size_t IosPacketLengthC(sizeof(IpHeader) + sizeof(IosTcpHeader));
static const size_t HelloPacketLengthC(sizeof(IpHeader) + sizeof(HelloTcpHeader) + HelloDataLengthC);
static const size_t HelloForHttpPacketLengthC(sizeof(IpHeader) + sizeof(HelloTcpHeader) + 155);

struct ContextInfo {
    bool isHelp;
    bool isDebug;
    bool isHttp;
    bool isDisableRST;
    bool isVerify;
    uint8_t ipHeaderTTL;
    string srcIp;
    string destIp;
    uint16_t srcPort;
    uint16_t destPort;
    uint32_t seqNo;                     // SYN（三次握手中的第一次）的SeqNumber
    uint32_t ackNo;                     // SYN（三次握手中的第一次）的AckNumber
    uint32_t synAckPacketSeqNo;         // SYN+ACK（三次握手中的第二次）的SeqNumber
    uint32_t lastSeqNo;
    uint32_t lastAckNo;
    uint32_t sendHelloDataLength;
    int sendSockFd;
    IpHeader *ipHeader;
    IosTcpHeader *iosTcpHeader;
    HelloTcpHeader *helloTcpHeader;
    sockaddr_in remoteAddr;
    uint8_t checksumBuffer[1024];
    uint8_t iosPacket[IosPacketLengthC];
    uint8_t helloPacket[HelloPacketLengthC];
    uint8_t helloForHttpPacket[HelloForHttpPacketLengthC];
    PsdHeader psdHeader;

    ContextInfo()
            : isHelp(false), isDebug(false), isHttp(true), isDisableRST(false), isVerify(false),
              ipHeaderTTL(64), srcIp(""), destIp(""), srcPort(0), destPort(0), seqNo(0),
              ackNo(0), synAckPacketSeqNo(0), lastSeqNo(0), lastAckNo(0), sendHelloDataLength(0), sendSockFd(-1),
              ipHeader(NULL),
              iosTcpHeader(NULL),
              helloTcpHeader(NULL) {
        bzero(&remoteAddr, sizeof(sockaddr_in));
        bzero(checksumBuffer, sizeof(checksumBuffer));
        bzero(iosPacket, sizeof(iosPacket));
        bzero(helloPacket, sizeof(helloPacket));
        bzero(helloForHttpPacket, sizeof(helloForHttpPacket));
        bzero(&psdHeader, sizeof(psdHeader));
    }

    ~ContextInfo() {
        if (isDisableRST) {
            enableRST(srcIp, destIp, srcPort, destPort);
        }

        if (-1 != sendSockFd) {
            if (0 > ::close(sendSockFd)) {
                cout << "Close raw socket failed." << endl;
            } else {
                cout << "Close raw socket" << endl;
            }
            sendSockFd = 0;
        }
    }
};

void showUsage() {
    cout << "Usage: Simulator Version: " << SIMULATOR_VERSION << endl;
    cout << "Options:" << endl;
    cout << " -h, --help                Print this message and exit." << endl;
    cout << " -p, --debug               Print debug log, optional, default false." << endl;
    cout << " -v, --verify              Verify the packets received from the server, optional, default false." << endl;
    cout << " -t, --tcp                 Using TCP packet of raw socket, for non-split tcp mode, optional, default false, will using HTTP packet." << endl;
    cout << " -x, --ttl                    Int, the TTL of SYNC packet, optional, default 64." << endl;
    cout << " -s                        String, the source IP addrss, must be specified." << endl;
    cout << " -d                        String, the destination IP addrss, must be specified." << endl;
    cout << " -l, --sport                  Int, the source port, optional, default will be automatically assigned." << endl;
    cout << " -r, --dport                  Int, the destination port, must be specified." << endl;
    cout << "Examples:" << endl;
    cout << " ./Simulator -s 1.0.0.1 -d 2.0.0.2 -l 6666 -r 8888 -t" << endl;
    cout << " ./Simulator -s 1.0.0.1 -d 2.0.0.2 --sport 6666--dport 8888 --tcp" << endl << endl;
}

int parseOpt(int argc, char *argv[], ContextInfo &context) {
    static struct option longOpts[] = {
            {"help",  no_argument,       NULL, 'h'},
            {"debug", no_argument,       NULL, 'p'},
            {"verify", no_argument,      NULL, 'v'},
            {"tcp",   no_argument,       NULL, 't'},
            {"ttl",   required_argument, NULL, 'x'},
            {"s",     required_argument, NULL, 's'},
            {"d",     required_argument, NULL, 'd'},
            {"sport", required_argument, NULL, 'l'},
            {"dport", required_argument, NULL, 'r'}
    };

    int optIndex = 0;
    for (;;) {
        optIndex = getopt_long(argc, argv, "hpvts:d:l:r:x:", longOpts, NULL);
        if (-1 == optIndex) {
            break;
        }

        switch (optIndex) {
            case 'p':
                context.isDebug = true;
                break;
            case 'v':
                context.isVerify = true;
                break;
            case 't':
                context.isHttp = false;
                break;
            case 'x':
                context.ipHeaderTTL = atoi(optarg);
                break;
            case 's':
                context.srcIp = string(optarg);
                break;
            case 'd':
                context.destIp = string(optarg);
                break;
            case 'l':
                context.srcPort = atoi(optarg);
                break;
            case 'r':
                context.destPort = atoi(optarg);
                break;
            case 'h':
            default:
                context.isHelp = true;
                break;
        }
    }

    if (context.isHelp) {
        return 0;
    }

    if (context.srcIp.empty()) {
        cout << "error: Source IP is empty." << endl;
        showUsage();
        return -1;
    }

    if (context.destIp.empty()) {
        cout << "error: Destination IP is empty." << endl;
        showUsage();
        return -1;
    }

    if (0 == context.srcPort) {
        int port = getAvaliablePort();
        if (-1 == port) {
            return -1;
        }

        context.srcPort = port;
    }

    if (0 == context.destPort) {
        cout << "error: Destination Port is 0." << endl;
        showUsage();
        return -1;
    }

    return 0;
}

uint16_t checkSum(uint16_t *buffer, int size) {
    //将变量放入寄存器, 提高处理效率.
    int len = size;
    //16bit
    uint16_t *p = buffer;
    //32bit
    uint32_t sum = 0;

    //16bit求和
    while (len >= 2) {
        sum += *(p++) & 0x0000ffff;
        len -= 2;
    }

    //最后的单字节直接求和
    if (len == 1) {
        sum += *((uint8_t *) p);
    }

    //高16bit与低16bit求和, 直到高16bit为0
    while ((sum & 0xffff0000) != 0) {
        sum = (sum >> 16) + (sum & 0x0000ffff);
    }
    return (uint16_t) (~sum);
}

int createRawSocket(ContextInfo &context) {
    context.sendSockFd = ::socket(AF_INET, SOCK_RAW, IPPROTO_TCP);
    if (context.sendSockFd < 0) {
        printf("create send socket failed\n");
        return -1;
    }

    int one = 1;
    if (::setsockopt(context.sendSockFd, IPPROTO_IP, IP_HDRINCL, &one, sizeof(one)) < 0) {   //定义套接字不添加IP首部，代码中手工添加
        printf("setsockopt IP_HDRINCL failed\n");
        return -1;
    }

    if (::setsockopt(context.sendSockFd, SOL_SOCKET, SO_REUSEADDR, &one, sizeof(one)) < 0) {   //端口复用
        printf("setsockopt SO_REUSEADDR failed\n");
        return -1;
    }

    context.remoteAddr.sin_family = AF_INET;
    context.remoteAddr.sin_addr.s_addr = inet_addr(context.destIp.c_str()); //设置接收方IP
    context.remoteAddr.sin_port = htons(context.destPort);
    return 0;
}

int sendIosSynPacket(ContextInfo &context) {
    bzero(context.iosPacket, sizeof(context.iosPacket));
    context.ipHeader = (IpHeader *) context.iosPacket;
    context.iosTcpHeader = (IosTcpHeader *) (context.iosPacket + sizeof(IpHeader));

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(IosPacketLengthC);
    context.ipHeader->ident = htons(0);
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = context.ipHeaderTTL;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.iosPacket,
                                          sizeof(IpHeader));  //计算IP首部的校验和，必须在其他字段都赋值后再赋值该字段，赋值前为0

    /*设置TCP首部*/
    uint16_t tcpHeaderLength = sizeof(IosTcpHeader);                                           // TCP fixed header length 20 + 24 byte Options
    context.iosTcpHeader->srcPort = htons(context.srcPort);
    context.iosTcpHeader->destPort = htons(context.destPort);
    context.iosTcpHeader->seq = htonl(context.seqNo);
    context.iosTcpHeader->ack = htons(context.ackNo);
    context.iosTcpHeader->headerLength = tcpHeaderLength / 4;
    context.iosTcpHeader->offset = 0;
    context.iosTcpHeader->flag = 0x02;                                                      //SYN置位
    context.iosTcpHeader->winSize = htons(65535);
    context.iosTcpHeader->surgent = htons(0);
    memcpy(context.iosTcpHeader->options, IosSynPacketTcpOptions, sizeof(IosSynPacketTcpOptions));

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str()); //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = IPPROTO_TCP;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength);

    bzero(&context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.iosTcpHeader, sizeof(IosTcpHeader));
    context.iosTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                              sizeof(context.psdHeader) + sizeof(IosTcpHeader));  //计算检验码

    int send = sendto(context.sendSockFd, context.iosPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("SYN packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = context.seqNo;
    context.lastAckNo = context.ackNo;
    printf("1 Client -----> SYN -----> Server ok, SeqNo=%u, AckNo=%u\n", context.seqNo, context.ackNo);
    return 0;
}

int recvAndCheckSynAckPacket(ContextInfo &context) {
    uint8_t synAckPacket[1024];
    while (1) {
        bzero(synAckPacket, sizeof(synAckPacket));
        //接收SYN+ACK报文
        int recvByte = recvfrom(context.sendSockFd, synAckPacket, 1024, 0, NULL, NULL);
        if (recvByte < 0) {
            printf("SYN+ACK packet recv failed, ret=%d\n", recvByte);
            return -1;
        }

        /*校验接收到的IP数据报，重新计算校验和，结果应为0*/
        uint8_t ipHeaderLength = synAckPacket[0];                                       //取出IP数据包的长度
        ipHeaderLength = (ipHeaderLength & 0x0f);                                       //IP首部长度字段只占该字节后四位
        ipHeaderLength *= 4;                                                            //四个字节为单位
        uint16_t ipTotalLength = ntohs(*((uint16_t *) (synAckPacket + 2)));             //获取IP数据报长度
        uint16_t tcpTotalLength = ipTotalLength - ipHeaderLength;                       //计算TCP数据报长度


        /*以下校验TCP报文，同样将伪首部和TCP报文放入buffer中*/
        bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
        for (int i = 0; i < 8; i++) {
            context.checksumBuffer[i] = synAckPacket[i + 12];                      //获取源IP和目的IP
        }
        context.checksumBuffer[8] = 0;                                             //伪首部的字段，可查阅资料
        context.checksumBuffer[9] = synAckPacket[9];                               //IP首部“上层协议”字段，即IPPROTO_TCP
        context.checksumBuffer[10] = 0;                                            //第10,11字节存储TCP报文长度，此处只考虑报文长度只用一个字节时，不会溢出，根据网络字节顺序存储
        uint8_t tcpHeaderLength = synAckPacket[32];                                //获取TCP报文长度
        tcpHeaderLength = tcpHeaderLength >> 4;                                    //因为TCP报文长度只占该字节的高四位，需要取出该四位的值
        tcpHeaderLength *= 4;                                                      //以四个字节为单位

        context.checksumBuffer[11] = tcpHeaderLength;                              //将TCP长度存入
        for (int i = 0; i < tcpTotalLength; i++) {                                 //buffer中加入TCP报文
            context.checksumBuffer[i + 12] = synAckPacket[i + ipHeaderLength];
        }

        /*检验收到的是否是SYN+ACK包，是否与上一个SYN请求包对应*/
        uint32_t synAckPacketSeqNo = ntohl(*((uint32_t *) (synAckPacket + ipHeaderLength + 4)));   //获取接收到的ACK+SYN包的序列号
        uint32_t synAckPacketAckNo = ntohl(*((int32_t *) (synAckPacket + ipHeaderLength + 8)));
        uint8_t synAckPacketFlag = synAckPacket[13 + ipHeaderLength];
        synAckPacketFlag = (synAckPacketFlag & 0x12);

        //判断是否为SYN+ACK包
        if (synAckPacketFlag != 0x12) {
            continue;
        }

        //判断是否是发送的SYN包的回应(SYN+ACK)
        if (synAckPacketAckNo != context.lastSeqNo + 1) {
            continue;
        }

        uint16_t ipHeaderChecksum = checkSum((uint16_t *) synAckPacket, ipHeaderLength);
        uint16_t tcpHeaderChecksum = checkSum((uint16_t *) context.checksumBuffer, 12 + tcpTotalLength);

        if (context.isDebug)
        {
            //将接受的IP数据报输出
            printf("Receive SYN+ACK packet %d bytes, hex stream:", recvByte);
            for (int i = 0; i < recvByte; i++) {
                if (i % 16 == 0) {
                    printf("\n\t");
                }
                printf("%02x ", synAckPacket[i]);
            }

            printf("\nReceive SYN+ACK packet info:\n\tipHeaderLength=%d, ipTotalLength=%d, ipHeaderChecksum=%d\n\t"
                "tcpHeaderLength=%d, tcpTotalLength=%d, tcpHeaderChecksum=%d\n\t"
                "synAckPacketSeqNo=%u, synAckPacketAckNo=%u\n",
                ipHeaderLength, ipTotalLength, ipHeaderChecksum, tcpHeaderLength,
                tcpTotalLength, tcpHeaderChecksum, synAckPacketSeqNo, synAckPacketAckNo);
        }

        context.synAckPacketSeqNo = synAckPacketSeqNo;
        context.lastSeqNo = synAckPacketSeqNo;
        context.lastAckNo = synAckPacketAckNo;
        printf("2 Client <----- SYN+ACK <----- Server ok, synAckPacketSeqNo=%u, synAckPacketAckNo=%u\n", synAckPacketSeqNo, synAckPacketAckNo);
        return 0;
    }
}

int sendIosAckPacket(ContextInfo &context) {
    bzero(context.iosPacket, sizeof(context.iosPacket));
    context.ipHeader = (IpHeader *) context.iosPacket;
    context.iosTcpHeader = (IosTcpHeader *) (context.iosPacket + sizeof(IpHeader));
    uint32_t seqNo = context.lastAckNo;
    uint32_t ackNo = context.lastSeqNo + 1;

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(IosPacketLengthC);
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.iosPacket,
                                          sizeof(IpHeader));  //计算IP首部的校验和，必须在其他字段都赋值后再赋值该字段，赋值前为0

    /*设置TCP首部*/
    uint16_t tcpHeaderLength = sizeof(IosTcpHeader);
    context.iosTcpHeader->srcPort = htons(context.srcPort);
    context.iosTcpHeader->destPort = htons(context.destPort);
    context.iosTcpHeader->seq = htonl(seqNo);
    context.iosTcpHeader->ack = ntohl(ackNo);
    context.iosTcpHeader->headerLength = tcpHeaderLength / 4;
    context.iosTcpHeader->offset = 0;
    context.iosTcpHeader->flag = 0x10;                                                         //ACK置位
    context.iosTcpHeader->winSize = htons(1026);
    context.iosTcpHeader->surgent = htons(0);
    memcpy(context.iosTcpHeader->options, IosSynPacketTcpOptions, sizeof(IosSynPacketTcpOptions));

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.iosTcpHeader, sizeof(IosTcpHeader));
    context.iosTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                              sizeof(context.psdHeader) + sizeof(IosTcpHeader));

    /*发送ACK报文段*/
    int send = sendto(context.sendSockFd, context.iosPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("ACK packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    printf("3 Client -----> ACK -----> Server ok, SeqNo=%u, AckNo=%u\n", seqNo, ackNo);
    return 0;
}

int sendHelloPacketForTcp(ContextInfo &context) {
    bzero(context.helloPacket, sizeof(context.helloPacket));
    context.ipHeader = (IpHeader *) context.helloPacket;
    context.helloTcpHeader = (HelloTcpHeader *) (context.helloPacket + sizeof(IpHeader));
    uint8_t *pData = context.helloPacket + sizeof(IpHeader) + sizeof(HelloTcpHeader);

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(HelloPacketLengthC);
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.helloPacket, sizeof(IpHeader));

    /*设置TCP首部*/
    uint32_t seqNo = context.lastSeqNo;
    uint32_t ackNo = context.lastAckNo;
    uint16_t tcpHeaderLength = sizeof(HelloTcpHeader);
    context.helloTcpHeader->srcPort = htons(context.srcPort);
    context.helloTcpHeader->destPort = htons(context.destPort);
    context.helloTcpHeader->seq = htonl(seqNo);
    context.helloTcpHeader->ack = ntohl(ackNo);
    context.helloTcpHeader->headerLength = tcpHeaderLength / 4;
    context.helloTcpHeader->offset = 0;
    context.helloTcpHeader->flag = 0x18;                                                         //PSH, ACK置位
    context.helloTcpHeader->winSize = htons(1026);
    context.helloTcpHeader->surgent = htons(0);
    memcpy(pData, HelloDataC.c_str(), HelloDataLengthC);                                         //应用层数据

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength + HelloDataLengthC);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.helloTcpHeader, sizeof(HelloTcpHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader) + sizeof(HelloTcpHeader), HelloDataC.c_str(),
           HelloDataLengthC);
    context.helloTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                                sizeof(context.psdHeader) + sizeof(HelloTcpHeader) + HelloDataLengthC);

    /*发送ACK报文段*/
    int send = sendto(context.sendSockFd, context.helloPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("Hello packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    context.sendHelloDataLength = HelloDataLengthC;
    printf("Client -----> Hello Server -----> Server ok, SeqNo=%u, AckNo=%u, HelloServerPacketDataLength=%u\n", seqNo, ackNo, HelloDataLengthC);
    return 0;
}

int sendHelloPacketForHttp(ContextInfo &context) {
    bzero(context.helloForHttpPacket, sizeof(context.helloForHttpPacket));
    context.ipHeader = (IpHeader *) context.helloForHttpPacket;
    context.helloTcpHeader = (HelloTcpHeader *) (context.helloForHttpPacket + sizeof(IpHeader));
    uint8_t *pData = context.helloForHttpPacket + sizeof(IpHeader) + sizeof(HelloTcpHeader);

    char httpData[256];
    bzero(httpData, sizeof(httpData));
    int httpDataLength = sprintf(httpData,
                                 "GET / HTTP/1.1\r\n"
                                 "Host: %s:%d\r\n"
                                 "User-Agent: Go-http-client/1.1\r\n"
                                 "Content-Length: %zu\r\n"
                                 "Clientsenddata: %s\r\n"
                                 "Accept-Encoding: gzip\r\n"
                                 "\r\n"
                                 "%s",
                                 context.destIp.c_str(),
                                 context.destPort,
                                 HelloDataLengthC,
                                 HelloDataC.c_str(),
                                 HelloDataC.c_str());

//    printf("HttpDataLength=%d, HttpData=\n%s\n", httpDataLength, httpData);

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(sizeof(IpHeader) + sizeof(HelloTcpHeader) + httpDataLength);
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.helloForHttpPacket, sizeof(IpHeader));

    /*设置TCP首部*/
    uint32_t seqNo = context.lastSeqNo;
    uint32_t ackNo = context.lastAckNo;
    uint16_t tcpHeaderLength = sizeof(HelloTcpHeader);
    context.helloTcpHeader->srcPort = htons(context.srcPort);
    context.helloTcpHeader->destPort = htons(context.destPort);
    context.helloTcpHeader->seq = htonl(seqNo);
    context.helloTcpHeader->ack = ntohl(ackNo);
    context.helloTcpHeader->headerLength = tcpHeaderLength / 4;
    context.helloTcpHeader->offset = 0;
    context.helloTcpHeader->flag = 0x18;                                             //PSH, ACK置位
    context.helloTcpHeader->winSize = htons(1026);
    context.helloTcpHeader->surgent = htons(0);
    memcpy(pData, httpData, httpDataLength);                                         //应用层数据

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength + httpDataLength);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.helloTcpHeader, sizeof(HelloTcpHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader) + sizeof(HelloTcpHeader), httpData,
           httpDataLength);
    context.helloTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                                sizeof(context.psdHeader) + sizeof(HelloTcpHeader) + httpDataLength);

    int send = sendto(context.sendSockFd, context.helloForHttpPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("Hello packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    context.sendHelloDataLength = httpDataLength;
    printf("Client -----> Hello Server -----> Server ok, SeqNo=%u, AckNo=%u, helloServerHttpDataLength=%u\n", seqNo, ackNo, httpDataLength);
    return 0;
}

int recvAndCheckHelloServerAckAndHelloClientPacket(ContextInfo &context) {
    uint8_t ackPacket[1024];
    uint32_t ackPacketSeqNo = 0;
    uint32_t ackPacketAckNo = 0;
    uint32_t helloClientPacketDataLength = 0;

    while (1) {
        if (!context.isVerify) {
            ackPacketSeqNo = context.lastAckNo;
            ackPacketAckNo = (context.lastSeqNo + context.sendHelloDataLength);
            helloClientPacketDataLength = 12;
        } else {
            bzero(ackPacket, sizeof(ackPacket));
            //接收Hello ACK报文
            int recvByte = recvfrom(context.sendSockFd, ackPacket, 1024, 0, NULL, NULL);
            if (recvByte < 0) {
                printf("packet recv failed, ret=%d\n", recvByte);
                return -1;
            }

            /*校验接收到的IP数据报，重新计算校验和，结果应为0*/
            uint8_t ipHeaderLength = ackPacket[0];                                          //取出IP数据包的长度
            ipHeaderLength = (ipHeaderLength & 0x0f);                                       //IP首部长度字段只占该字节后四位
            ipHeaderLength *= 4;                                                            //四个字节为单位
            uint16_t ipTotalLength = ntohs(*((uint16_t *) (ackPacket + 2)));                //获取IP数据报长度
            uint16_t tcpTotalLength = ipTotalLength - ipHeaderLength;                       //计算TCP数据报长度

            /*以下校验TCP报文，同样将伪首部和TCP报文放入buffer中*/
            bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
            for (int i = 0; i < 8; i++) {
                context.checksumBuffer[i] = ackPacket[i + 12];                          //获取源IP和目的IP
            }
            context.checksumBuffer[8] = 0;                                              //伪首部的字段，可查阅资料
            context.checksumBuffer[9] = ackPacket[9];                                   //IP首部“上层协议”字段，即IPPROTO_TCP
            context.checksumBuffer[10] = 0;                                             //第10,11字节存储TCP报文长度，此处只考虑报文长度只用一个字节时，不会溢出，根据网络字节顺序存储
            uint8_t tcpHeaderLength = ackPacket[32];                                    //获取TCP报文长度
            tcpHeaderLength = tcpHeaderLength >> 4;                                     //因为TCP报文长度只占该字节的高四位，需要取出该四位的值
            tcpHeaderLength *= 4;                                                       //以四个字节为单位

            context.checksumBuffer[11] = tcpHeaderLength;                               //将TCP长度存入
            for (int i = 0; i < tcpTotalLength; i++) {                                  //buffer中加入TCP报文
                context.checksumBuffer[i + 12] = ackPacket[i + ipHeaderLength];
            }

            /*检验收到的是否是SYN+ACK包，是否与上一个SYN请求包对应*/
            ackPacketSeqNo = ntohl(*((uint32_t *) (ackPacket + ipHeaderLength + 4)));   //获取接收到的SYN包的序列号
            ackPacketAckNo = ntohl(*((int32_t *) (ackPacket + ipHeaderLength + 8)));
            uint8_t flag = ackPacket[13 + ipHeaderLength];
            uint8_t tcpFlag = (flag & 0x18);
            uint8_t httpAckFlag = (flag & 0x10);

            // Hello Client packet
            if (!context.isHttp) {
                if (tcpFlag != 0x18) {
                    continue;
                }

                if (ackPacketSeqNo != context.lastAckNo) {
                    continue;
                }

                if (ackPacketAckNo != (context.lastSeqNo + context.sendHelloDataLength)) {
                    continue;
                }
            } else {
                if (tcpFlag != 0x10) {
                    continue;
                }

                if (ackPacketSeqNo != context.lastAckNo) {
                    continue;
                }

                if (ackPacketAckNo != (context.lastSeqNo + context.sendHelloDataLength)) {
                    continue;
                }
            }

            uint16_t ipHeaderChecksum = checkSum((uint16_t *) ackPacket, ipHeaderLength);
            uint16_t tcpHeaderChecksum = checkSum((uint16_t *) context.checksumBuffer, 12 + tcpTotalLength);

            if (context.isDebug)
            {
                //将接受的IP数据报输出
                printf("Receive Hello ACK packet %d bytes, hex stream:", recvByte);
                for (int i = 0; i < recvByte; i++) {
                    if (i % 16 == 0) {
                        printf("\n\t");
                    }
                    printf("%02x ", ackPacket[i]);
                }

                printf("\nReceive Hello ACK packet info:\n\tipHeaderLength=%d, ipTotalLength=%d, ipHeaderChecksum=%d\n\t"
                    "tcpHeaderLength=%d, tcpTotalLength=%d, tcpHeaderChecksum=%d\n\t"
                    "ackPacketSeqNo=%u, ackPacketAckNo=%u\n",
                    ipHeaderLength, ipTotalLength, ipHeaderChecksum, tcpHeaderLength,
                    tcpTotalLength, tcpHeaderChecksum, ackPacketSeqNo, ackPacketAckNo);
            }

            helloClientPacketDataLength = tcpTotalLength - tcpHeaderLength;
        }

        context.lastSeqNo = ackPacketSeqNo;
        context.lastAckNo = ackPacketAckNo;
        printf("Client <----- Hello Client <----- Server ok, SeqNo=%u, AckNo=%u, HelloClientPacketDataLength=%u\n", ackPacketSeqNo, ackPacketAckNo, helloClientPacketDataLength);
        return 0;
    }
}

int sendHelloClientAckPacketToServer(ContextInfo &context) {
    bzero(context.helloPacket, sizeof(context.helloPacket));
    context.ipHeader = (IpHeader *) context.helloPacket;
    context.helloTcpHeader = (HelloTcpHeader *) (context.helloPacket + sizeof(IpHeader));

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(sizeof(IpHeader) + sizeof(HelloTcpHeader));
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x0);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.helloPacket, sizeof(IpHeader));

    /*设置TCP首部*/
    uint32_t seqNo = context.lastAckNo;
    uint32_t ackNo = 0;
    if (!context.isHttp) {
        ackNo = context.lastSeqNo + context.sendHelloDataLength;
    } else {
        ackNo = context.lastSeqNo + 130;
    }

    uint16_t tcpHeaderLength = sizeof(HelloTcpHeader);
    context.helloTcpHeader->srcPort = htons(context.srcPort);
    context.helloTcpHeader->destPort = htons(context.destPort);
    context.helloTcpHeader->seq = htonl(seqNo);
    context.helloTcpHeader->ack = htonl(ackNo);
    context.helloTcpHeader->headerLength = tcpHeaderLength / 4;
    context.helloTcpHeader->offset = 0;
    context.helloTcpHeader->flag = 0x10;                                                         //ACK置位
    context.helloTcpHeader->winSize = htons(1026);
    context.helloTcpHeader->surgent = htons(0);

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.helloTcpHeader, sizeof(HelloTcpHeader));

    context.helloTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                                sizeof(context.psdHeader) + sizeof(HelloTcpHeader));

    /*发送Final报文段*/
    int send = sendto(context.sendSockFd, context.helloPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("Hello Client Ack packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    printf("Client -----> Hello Client ACK -----> Server ok, SeqNo=%u, AckNo=%u\n", seqNo, ackNo);
    return 0;
}

int sendFinalPacket(ContextInfo &context) {
    bzero(context.helloPacket, sizeof(context.helloPacket));
    context.ipHeader = (IpHeader *) context.helloPacket;
    context.helloTcpHeader = (HelloTcpHeader *) (context.helloPacket + sizeof(IpHeader));

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(sizeof(IpHeader) + sizeof(HelloTcpHeader));
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.helloPacket, sizeof(IpHeader));

    /*设置TCP首部*/
    uint32_t seqNo = context.lastSeqNo;
    uint32_t ackNo = context.lastAckNo;
    uint16_t tcpHeaderLength = sizeof(HelloTcpHeader);
    context.helloTcpHeader->srcPort = htons(context.srcPort);
    context.helloTcpHeader->destPort = htons(context.destPort);
    context.helloTcpHeader->seq = htonl(seqNo);
    context.helloTcpHeader->ack = ntohl(ackNo);
    context.helloTcpHeader->headerLength = tcpHeaderLength / 4;
    context.helloTcpHeader->offset = 0;
    context.helloTcpHeader->flag = 0x11;                                                         //FIN, ACK置位
    context.helloTcpHeader->winSize = htons(1026);
    context.helloTcpHeader->surgent = htons(0);

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.helloTcpHeader, sizeof(HelloTcpHeader));

    context.helloTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                                sizeof(context.psdHeader) + sizeof(HelloTcpHeader));

    /*发送Final报文段*/
    int send = sendto(context.sendSockFd, context.helloPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("Final packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    printf("1 Client -----> FIN+ACK -----> Server ok, SeqNo=%u, AckNo=%u\n", seqNo, ackNo);
    return 0;
}

int recvAndCheckFinalAckPacket(ContextInfo &context) {
    uint8_t finalAckPacket[1024];
    uint32_t finalAckPacketSeqNo = 0;
    uint32_t finalAckPacketAckNo = 0;

    while (1) {
        if (!context.isVerify) {
            finalAckPacketSeqNo = context.lastAckNo;
            finalAckPacketAckNo = (context.lastSeqNo + 1);
        } else {
            bzero(finalAckPacket, sizeof(finalAckPacket));
            //接收FIN+ACK报文
            int recvByte = recvfrom(context.sendSockFd, finalAckPacket, 1024, 0, NULL, NULL);
            if (recvByte < 0) {
                printf("FIN+ACK packet recv failed, ret=%d\n", recvByte);
                return -1;
            }

            /*校验接收到的IP数据报，重新计算校验和，结果应为0*/
            uint8_t ipHeaderLength = finalAckPacket[0];                                          //取出IP数据包的长度
            ipHeaderLength = (ipHeaderLength & 0x0f);                                       //IP首部长度字段只占该字节后四位
            ipHeaderLength *= 4;                                                            //四个字节为单位
            uint16_t ipTotalLength = ntohs(*((uint16_t *) (finalAckPacket + 2)));                //获取IP数据报长度
            uint16_t tcpTotalLength = ipTotalLength - ipHeaderLength;                       //计算TCP数据报长度


            /*以下校验TCP报文，同样将伪首部和TCP报文放入buffer中*/
            bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
            for (int i = 0; i < 8; i++) {
                context.checksumBuffer[i] = finalAckPacket[i + 12];                          //获取源IP和目的IP
            }
            context.checksumBuffer[8] = 0;                                              //伪首部的字段，可查阅资料
            context.checksumBuffer[9] = finalAckPacket[9];                                   //IP首部“上层协议”字段，即IPPROTO_TCP
            context.checksumBuffer[10] = 0;                                             //第10,11字节存储TCP报文长度，此处只考虑报文长度只用一个字节时，不会溢出，根据网络字节顺序存储
            uint8_t tcpHeaderLength = finalAckPacket[32];                                    //获取TCP报文长度
            tcpHeaderLength = tcpHeaderLength >> 4;                                     //因为TCP报文长度只占该字节的高四位，需要取出该四位的值
            tcpHeaderLength *= 4;                                                       //以四个字节为单位

            context.checksumBuffer[11] = tcpHeaderLength;                               //将TCP长度存入
            for (int i = 0; i < tcpTotalLength; i++) {                                  //buffer中加入TCP报文
                context.checksumBuffer[i + 12] = finalAckPacket[i + ipHeaderLength];
            }

            /*检验收到的是否是FIN+ACK包，是否与上一个FIN+ACK对应*/
            finalAckPacketSeqNo = ntohl(*((uint32_t *) (finalAckPacket + ipHeaderLength + 4)));   //获取接收到的SYN包的序列号
            finalAckPacketAckNo = ntohl(*((int32_t *) (finalAckPacket + ipHeaderLength + 8)));
            uint8_t finalAckPacketFlag = finalAckPacket[13 + ipHeaderLength];
            finalAckPacketFlag = (finalAckPacketFlag & 0x11);


            if (finalAckPacketFlag != 0x11) {
//            printf("This is not FIN+ACK packet, ackPacketFlag=%02x\n", finalAckPacketFlag);
//            usleep(30);
                continue;
            }

            if (finalAckPacketSeqNo != context.lastAckNo) {
                continue;
            }

            if (finalAckPacketAckNo != context.lastSeqNo + 1) {
//            printf("This is not match an FIN+ACK with send Hello\n");
//            usleep(30);
                continue;
            }

            uint16_t ipHeaderChecksum = checkSum((uint16_t *) finalAckPacket, ipHeaderLength);
            uint16_t tcpHeaderChecksum = checkSum((uint16_t *) context.checksumBuffer, 12 + tcpTotalLength);

            if (context.isDebug)
            {
                //将接受的IP数据报输出
                printf("Receive FIN+ACK packet %d bytes, hex stream:", recvByte);
                for (int i = 0; i < recvByte; i++) {
                    if (i % 16 == 0) {
                        printf("\n\t");
                    }
                    printf("%02x ", finalAckPacket[i]);
                }

                printf("\nReceive FIN+ACK packet info:\n\tipHeaderLength=%d, ipTotalLength=%d, ipHeaderChecksum=%d\n\t"
                    "tcpHeaderLength=%d, tcpTotalLength=%d, tcpHeaderChecksum=%d\n\t"
                    "finalAckPacketSeqNo=%u, finalAckPacketAckNo=%u\n",
                    ipHeaderLength, ipTotalLength, ipHeaderChecksum, tcpHeaderLength,
                    tcpTotalLength, tcpHeaderChecksum, finalAckPacketSeqNo, finalAckPacketAckNo);
            }
        }

        context.lastSeqNo = finalAckPacketSeqNo;
        context.lastAckNo = finalAckPacketAckNo;
        printf("2 Client <----- FIN+ACK <----- Server ok, SeqNo=%u, AckNo=%u\n", finalAckPacketSeqNo, finalAckPacketAckNo);
        return 0;
    }
}

int sendLastAckPacket(ContextInfo &context) {
    bzero(context.helloPacket, sizeof(context.helloPacket));
    context.ipHeader = (IpHeader *) context.helloPacket;
    context.helloTcpHeader = (HelloTcpHeader *) (context.helloPacket + sizeof(IpHeader));

    /*设置IP首部*/
    context.ipHeader->headerLength = 5;
    context.ipHeader->ipVersion = 4;
    context.ipHeader->tos = 0;
    context.ipHeader->totalLength = htons(sizeof(IpHeader) + sizeof(HelloTcpHeader));
    context.ipHeader->ident = htons(13543); // random()
    context.ipHeader->off = htons(0x4000);
    context.ipHeader->ttl = 64;
    context.ipHeader->proto = IPPROTO_TCP;
    context.ipHeader->srcIp = inet_addr(context.srcIp.c_str());
    context.ipHeader->destIp = inet_addr(context.destIp.c_str());
    context.ipHeader->checksum = checkSum((uint16_t *) context.helloPacket, sizeof(IpHeader));

    /*设置TCP首部*/
    uint32_t seqNo = context.lastAckNo;
    uint32_t ackNo = context.lastSeqNo + 1;
    uint16_t tcpHeaderLength = sizeof(HelloTcpHeader);
    context.helloTcpHeader->srcPort = htons(context.srcPort);
    context.helloTcpHeader->destPort = htons(context.destPort);
    context.helloTcpHeader->seq = htonl(seqNo);
    context.helloTcpHeader->ack = ntohl(ackNo);
    context.helloTcpHeader->headerLength = tcpHeaderLength / 4;
    context.helloTcpHeader->offset = 0;
    context.helloTcpHeader->flag = 0x10;                                                         //ACK置位
    context.helloTcpHeader->winSize = htons(1026);
    context.helloTcpHeader->surgent = htons(0);

    /*设置tcp伪首部，用于计算TCP报文段校验和*/
    bzero(&context.psdHeader, sizeof(context.psdHeader));
    context.psdHeader.srcAddr = inet_addr(context.srcIp.c_str());                   //源IP地址
    context.psdHeader.destAddr = inet_addr(context.destIp.c_str());                 //目的IP地址
    context.psdHeader.mbz = 0;
    context.psdHeader.proto = 6;
    context.psdHeader.tcpHeaderAndDataLength = htons(tcpHeaderLength);

    bzero(context.checksumBuffer, sizeof(context.checksumBuffer));
    memcpy(context.checksumBuffer, &context.psdHeader, sizeof(context.psdHeader));
    memcpy(context.checksumBuffer + sizeof(context.psdHeader), context.helloTcpHeader, sizeof(HelloTcpHeader));

    context.helloTcpHeader->checksum = checkSum((uint16_t *) context.checksumBuffer,
                                                sizeof(context.psdHeader) + sizeof(HelloTcpHeader));

    /*发送Final报文段*/
    int send = sendto(context.sendSockFd, context.helloPacket, htons(context.ipHeader->totalLength), 0,
                      (sockaddr *) &context.remoteAddr,
                      sizeof(context.remoteAddr));
    if (send < 0) {
        printf("Final packet send failed, ret=%d\n", send);
        return -1;
    }

    context.lastSeqNo = seqNo;
    context.lastAckNo = ackNo;
    printf("3 Client -----> ACK -----> Server ok, SeqNo=%u, AckNo=%u\n", seqNo, ackNo);
    return 0;
}

// 在Linux系统中，当发出SYN包并接收到SYN+ACK包之后，Client端的Kernel会发一个RST包给Server端，
// 因为Client端的Kernel不认识Server端发回的ACK包，需要使用iptables屏蔽掉RST，禁止RST发出
// iptables -A OUTPUT -p tcp -s 1.0.0.1 -d 1.0.0.1 --sport 6666 --dport 8000 --tcp-flags ALL RST -j DROP
// iptables -D OUTPUT -p tcp -s 1.0.0.1 -d 1.0.0.1 --sport 6666 --dport 8000 --tcp-flags ALL RST -j DROP
int disableRST(ContextInfo &context) {
    char buf[256];
    sprintf(buf, "iptables -A OUTPUT -p tcp -s %s -d %s --sport %d --dport %d --tcp-flags ALL RST -j DROP",
            context.srcIp.c_str(), context.destIp.c_str(), context.srcPort, context.destPort);

    if (0 > system(buf)) {
        printf("Disable RST packet failed\n");
        return -1;
    } else {
        context.isDisableRST = true;
        printf("Disable RST packet ok, cmd=%s\n", buf);
    }

    return 0;
}

int main(int argc, char *argv[]) {
    ContextInfo context;
    if (-1 == parseOpt(argc, argv, context)) {
        return -1;
    }

    if (context.isHelp) {
        showUsage();
        return 0;
    }

    cout << "Using IOS TCP header options of the raw socket, that is a " << (context.isHttp ? "HTTP traffic, " : "TCP traffic, ") 
         << (context.isVerify ? "verify the packets received from the server, " : "do not verify the packets received from the server, ") 
         << "localAddress=" << context.srcIp << ":" << context.srcPort << ", remoteAddress=" << context.destIp << ":" << context.destPort << endl;

    if (-1 == createRawSocket(context)) {
        return -1;
    }

    disableRST(context);
    cout << "****************************************************************************" << endl;

    if (-1 == sendIosSynPacket(context)) {
        return -1;
    }

    if (-1 == recvAndCheckSynAckPacket(context)) {
        return -1;
    }

    if (-1 == sendIosAckPacket(context)) {
        return -1;
    }

    printf("     ---------------------------------     \n");

    if (context.isHttp) {
        if (-1 == sendHelloPacketForHttp(context)) {
            return -1;
        }
    } else {
        if (-1 == sendHelloPacketForTcp(context)) {
            return -1;
        }
    }

    // HelloServerACK packet and HelloClient packet
    if (-1 == recvAndCheckHelloServerAckAndHelloClientPacket(context)) {
        return -1;
    }

    sleep(2);

    // HelloClientAck
    if (-1 == sendHelloClientAckPacketToServer(context)) {
        return -1;
    }

    sleep(1);

    printf("     ---------------------------------     \n");
    if (-1 == sendFinalPacket(context)) {
        return -1;
    }

    sleep(2);
    if (-1 == recvAndCheckFinalAckPacket(context)) {
        return -1;
    }

    if (-1 == sendLastAckPacket(context)) {
        return -1;
    }

    cout << "****************************************************************************" << endl;

    sleep(1);
    cout << "successful" << endl;
    return 0;
}