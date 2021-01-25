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
std::string SERV_IP=("0.0.0.0");
uint32_t SERV_PORT=8001;
std::string CLT_IP=("0.0.0.0");
uint32_t CLT_PORT=8989;

void showUsage() {
    cout << "RstFinTool Usage:" << endl;
    cout << " -h, --help                Print this message and exit." << endl;
    cout << " -s, --server              Run the RstFinTool as tcp server. default run as client." << endl;
    cout << " -r, --rst                 The RstFinTool send a RST packes to the client after 3-handshakes complete, that is select when unspecified." << endl;
    cout << " -f, --fin                 The RstFinTool send a FIN packes to the client after 3-handshakes complete." << endl;
    cout << " -i, --ip                  String, the server  IP addrss, must be specified. default as 0.0.0.0." << endl;
    cout << " -p, --port                Int, the server port, must be specified. default as 8001." << endl;
    cout << " -c, --sip                 String, the client IP addrss, must be specified. default as 0.0.0.0." << endl;
    cout << " -d, --sport               Int, the client port, must be specified. default as 8989." << endl;
    cout << "RstFinTool Examples:" << endl;
    cout << " ./RstFinTool -s -r -i 127.0.0.1 -p 6666  # Run RstFinTool as RST server listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool -s -f -i 127.0.0.1 -p 6666  # Run RstFinTool as FIN server listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool -r -i 127.0.0.1 -p 6666     # Send data to RST server, that listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool -f -i 127.0.0.1 -p 6666     # Send data to FIN server, that listen on 127.0.0.1:6666" << endl;
    cout << " ./RstFinTool -f -c 127.0.0.1 -d 8888 -i 127.0.0.1 -p 6666     # the Client bind on 127.0.0.1:8888 and send data to FIN server, that listen on 127.0.0.1:6666" << endl;
}

int parseOpt(int argc, char *argv[]) {
   static struct option longOpts[] = {
      {"help", no_argument, NULL, 'h'},
      {"server", no_argument, NULL, 's'},
      {"rst", no_argument, NULL, 'r'},
      {"fin", no_argument, NULL, 'f'},
      {"ip", required_argument, NULL, 'i'},
      {"port", required_argument, NULL, 'p'},
      {"sip", optional_argument, NULL, 'c'},
      {"sport", optional_argument, NULL, 'd'}
   };

   int optIndex = 0;
   for (;;) 
   {
      optIndex = getopt_long(argc, argv, "hsrf:i:p:c:d:", longOpts, NULL);
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
         case 'i':
               SERV_IP = std::string(optarg);
               break;
         case 'p':
               SERV_PORT = atoi(optarg);
               break;
         case 'c':
               CLT_IP = std::string(optarg);
               break;
         case 'd':
               CLT_PORT = atoi(optarg);
               break;
         case 'h':
         default:
               IS_HELP = true;
               break;
      }
   }

   return 0;
}

int handleAccept(int fd)
{
   if (IS_RST)
   {
      char pcContent[4096];
      ::bzero(pcContent, sizeof(pcContent));
      ::read(fd,pcContent, 4096);
      printf("close the fd=%d for RST\n", fd);
   }

   if (IS_FIN)
   {
      char pcContent[128];
      ::bzero(pcContent, sizeof(pcContent));
      ::read(fd,pcContent, 128);
      printf("close the fd=%d for FIN\n", fd);
   }
  
   ::close(fd);
   return 0;
}

int startServer()
{
   int listen_fd(-1), client_fd(-1);
   struct sockaddr_in listen_addr, client_addr;
   socklen_t len = sizeof(struct sockaddr_in);
   listen_fd = ::socket(AF_INET, SOCK_STREAM, 0);
   if(listen_fd == -1)
   {
      perror("socket failed");
      return -1;
   }

   ::bzero(&listen_addr,sizeof(listen_addr));
   listen_addr.sin_family = AF_INET;
   listen_addr.sin_addr.s_addr = htonl(INADDR_ANY);
   listen_addr.sin_port = htons(SERV_PORT);
   ::bind(listen_fd,(struct sockaddr *)&listen_addr, len);
   ::listen(listen_fd, 9999);
   printf("tcp server listen on 0.0.0.0:%d, type=%s, fd=%d\n", SERV_PORT, IS_RST ? "RST" : "FIN", listen_fd);

   while(1)
   {
      client_fd = ::accept(listen_fd, (struct sockaddr*)&client_addr, &len);
      if(client_fd == -1)
      {
         perror("tcp server accpet fail");
         return -1;
      }
      else
      {
         int port = ntohs(client_addr.sin_port);
         char ipDotDec[INET_ADDRSTRLEN];
         ::bzero(ipDotDec, sizeof(ipDotDec));
         ::inet_ntop(AF_INET, &(client_addr.sin_addr), ipDotDec, sizeof(ipDotDec));
         printf("**************************************************\n");
         printf("connection fd=%d, client address=%s:%d\n", client_fd, ipDotDec, port);
      }

      std::thread c(handleAccept, client_fd);
      c.detach();
   }

   ::close(listen_fd);
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
   s_addr.sin_addr.s_addr = inet_addr(CLT_IP.c_str());
   s_addr.sin_port = htons(CLT_PORT);
   if (bind(send_fd, (struct sockaddr *) &s_addr, sizeof(s_addr)) < 0) 
   {
      perror("bind failed");
      return -1;
   }

   bzero(&d_addr, sizeof(d_addr));
   d_addr.sin_family = AF_INET;
   inet_pton(AF_INET, SERV_IP.c_str(), &d_addr.sin_addr);
   d_addr.sin_port = htons(SERV_PORT);
   if(connect(send_fd, (struct sockaddr*)&d_addr, len) == -1) 
   { 
      perror("connect fail");
      return -1;
   }

   if (IS_RST)
   {
      // IP fragmentation
      char pcContent[5000]={0};
      ::write(send_fd,pcContent,5000);
      printf("send data for RST\n");
   }

   if (IS_FIN)
   {
      char pcContent[16]={0};
      ::write(send_fd,pcContent,16);
      printf("send data for FIN\n");
   }

   ::sleep(1);
   ::close(send_fd);
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
