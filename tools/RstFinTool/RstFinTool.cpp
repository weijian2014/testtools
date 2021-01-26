#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <thread>
#include <string>
#include <getopt.h>
#include <iostream>

/*

Build:
   g++ -std=c++11 -pthread RstFinTool.cpp -o RstFinTool

Function:
   Simulate the TCP server send a RST/FIN packet to client after 3-handshakes complete.

*/

using namespace std;

bool IS_HELP=false;
bool IS_SERVER=false;
bool IS_RST=true;
bool IS_FIN=false;
string SRC_IP=("0.0.0.0");
uint32_t SRC_PORT=6666;
string DST_IP=("0.0.0.0");
uint32_t DST_PORT=8888;

void showUsage() {
    cout << "RstFinTool: simulate the TCP server send a RST/FIN packet to client after 3-handshakes complete." << endl;
    cout << "1) Usage:" << endl;
    cout << " -h, --help                Print this message and exit." << endl;
    cout << " -s, --server              Run the RstFinTool as tcp server, default run as client." << endl;
    cout << " -r, --rst                 The RstFinTool send a RST packes to the client after 3-handshakes complete which dafault server type" << endl;
    cout << " -f, --fin                 The RstFinTool send a FIN packes to the client after 3-handshakes complete." << endl;
    cout << " -a, --sip                 String, the source IP addrss, default is 0.0.0.0." << endl;
    cout << " -b, --sport               Int, the source port, default is 6666 as server, default is random as client." << endl;
    cout << " -c, --dip                 String, the destination IP addrss, default is 0.0.0.0." << endl;
    cout << " -d, --dport               Int, the destination port, default is 8888." << endl;
    cout << "2) Examples:" << endl;
    cout << " ./RstFinTool --server --sip 127.0.0.1 --sport 6666 --rst  # Run RstFinTool as RST server which listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool --server --sip 127.0.0.1 --sport 6666 --fin  # Run RstFinTool as FIN server which listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool --sip 127.0.0.1 --sport 8888 --dip 127.0.0.1 --dport 6666 --fin" << endl;
    cout << " \t\t\t# The client[127.0.0.1:8888] send data to FIN server[127.0.0.1:6666]" << endl;
}

int parseOpt(int argc, char *argv[]) {
   static struct option longOpts[] = {
      {"help", no_argument, NULL, 'h'},
      {"server", no_argument, NULL, 's'},
      {"rst", no_argument, NULL, 'r'},
      {"fin", no_argument, NULL, 'f'},
      {"sip", required_argument, NULL, 'a'},
      {"sport", required_argument, NULL, 'b'},
      {"dip", required_argument, NULL, 'c'},
      {"dport", required_argument, NULL, 'd'}
   };

   int optIndex = 0;
   for (;;) 
   {
      optIndex = getopt_long(argc, argv, "hsrf:a:b:c:d:", longOpts, NULL);
      if (-1 == optIndex)
      {
         break;
      }

      switch (optIndex) {
         case 's':
               IS_SERVER = true;
               break;
         case 'r':
               IS_RST = true;
               IS_FIN = false;
               break;
         case 'f':
               IS_RST = false;
               IS_FIN = true;
               break;
         case 'a':
               SRC_IP = string(optarg);
               break;
         case 'b':
               SRC_PORT = atoi(optarg);
               break;
         case 'c':
               DST_IP = string(optarg);
               break;
         case 'd':
               DST_PORT = atoi(optarg);
               break;
         case 'h':
         default:
               IS_HELP = true;
               break;
      }
   }

   if (!IS_SERVER)
   {
      // At this point, you can reach for the port 0 trick: 
      // on both Windows and Linux, if you bind a socket to port 0, the kernel will assign it a free port number somewhere above 1024.
      SRC_PORT=0;
   }

   return 0;
}

int handleAccept(int fd)
{
   if (IS_RST)
   {
      char pcContent[4096];
      bzero(pcContent, sizeof(pcContent));
      read(fd,pcContent, 4096);
      printf("send RST packet to client ok\n\n");
   }

   if (IS_FIN)
   {
      char pcContent[128];
      bzero(pcContent, sizeof(pcContent));
      read(fd,pcContent, 128);
      printf("send FIN packet to client ok\n\n");
   }
  
   close(fd);
   return 0;
}

int startServer()
{
   int listen_fd(-1), client_fd(-1);
   struct sockaddr_in listen_addr, client_addr;
   socklen_t len = sizeof(struct sockaddr_in);
   listen_fd = socket(AF_INET, SOCK_STREAM, 0);
   if(listen_fd == -1)
   {
      perror("socket failed");
      return -1;
   }

   bzero(&listen_addr,sizeof(listen_addr));
   listen_addr.sin_family = AF_INET;
   listen_addr.sin_addr.s_addr = inet_addr(SRC_IP.c_str());    // INADDR_ANY
   listen_addr.sin_port = htons(SRC_PORT);
   bind(listen_fd,(struct sockaddr *)&listen_addr, len);
   listen(listen_fd, 9999);
   printf("tcp server listen on %s:%d, type=%s, fd=%d\n", SRC_IP.c_str(), SRC_PORT, IS_RST ? "RST" : "FIN", listen_fd);

   while(1)
   {
      client_fd = accept(listen_fd, (struct sockaddr*)&client_addr, &len);
      if(client_fd == -1)
      {
         perror("tcp server accpet fail");
         return -1;
      }
      else
      {
         int port = ntohs(client_addr.sin_port);
         char ipDotDec[INET_ADDRSTRLEN];
         bzero(ipDotDec, sizeof(ipDotDec));
         inet_ntop(AF_INET, &(client_addr.sin_addr), ipDotDec, sizeof(ipDotDec));
         printf("*********client[%s:%d] connected, fd=%d*********\n", ipDotDec, port, client_fd);
      }

      thread t(handleAccept, client_fd);
      t.detach();
   }

   close(listen_fd);
}

int startClient()
{
   int send_fd;
   send_fd = socket(AF_INET, SOCK_STREAM, 0);
   if(send_fd == -1) 
   { 
      perror("socket failed");
      return -1;
   }

   struct sockaddr_in s_addr, d_addr;
   socklen_t len = sizeof(s_addr);
   s_addr.sin_family = AF_INET;
   s_addr.sin_addr.s_addr = inet_addr(SRC_IP.c_str());
   s_addr.sin_port = htons(SRC_PORT);
   if (bind(send_fd, (struct sockaddr *) &s_addr, sizeof(s_addr)) < 0)
   {
      perror("bind failed");
      return -1;
   }

   bzero(&d_addr, sizeof(d_addr));
   d_addr.sin_family = AF_INET;
   inet_pton(AF_INET, DST_IP.c_str(), &d_addr.sin_addr);
   d_addr.sin_port = htons(DST_PORT);
   if(connect(send_fd, (struct sockaddr*)&d_addr, len) == -1) 
   {
      perror("connect fail");
      return -1;
   }

   if (IS_RST)
   {
      // IP fragmentation
      char pcContent[2050]={0};
      write(send_fd, pcContent, 2050);
      printf("client[%s:%d] send data to server[%s:%d] for RST\n", SRC_IP.c_str(), SRC_PORT, DST_IP.c_str(), DST_PORT);
   }

   if (IS_FIN)
   {
      char pcContent[16]={0};
      write(send_fd, pcContent, 16);
      printf("client[%s:%d] send data to server[%s:%d] for FIN\n", SRC_IP.c_str(), SRC_PORT, DST_IP.c_str(), DST_PORT);
   }

   sleep(3);
   close(send_fd);
   return 0;
}

int main(int argc, char** argv) 
{
   if (0 != parseOpt(argc, argv))
   {
      printf("parse option fail\n");
      return -1;
   }

   if (IS_HELP)
   {
      showUsage();
      return 0;
   }

   if (IS_SERVER)
   {
      startServer();
   }
   else
   {
      startClient();
   }
   
   return 0;
}
