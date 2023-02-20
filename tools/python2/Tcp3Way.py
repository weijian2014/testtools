#!/usr/bin/python

from scapy.all import *
import logging
import argparse
import os
import socket
logging.getLogger('scapy.runtime').setLevel(logging.ERROR)

def disable_reset(src_ip, dest_ip, src_port, dest_port, version):
    if (6 == version):
        cmd = '''ip6tables -A OUTPUT -p tcp -d {0} --tcp-flags RST RST -j DROP'''.format(dest_ip)
    else:
        cmd = '''iptables -A OUTPUT -p tcp -d {0} --tcp-flags RST RST -j DROP'''.format(dest_ip)
    # print('Disable RST, cmd={0}'.format(cmd))
    return os.system(cmd)

def enable_reset(src_ip, dest_ip, src_port, dest_port, version):
    cmd = ""
    if (6 == version):
        cmd = '''ip6tables -D OUTPUT -p tcp -d {0} --tcp-flags RST RST -j DROP'''.format(dest_ip)
    else:
        cmd = '''iptables -D OUTPUT -p tcp -d {0} --tcp-flags RST RST -j DROP'''.format(dest_ip)
    # print('Enable RST, cmd={0}'.format(cmd))
    return os.system(cmd)

def get_port():
    pscmd = "netstat -ntl |grep -v Active| grep -v Proto|awk '{print $4}'|awk -F: '{print $NF}'"
    procs = os.popen(pscmd).read()
    procarr = procs.split("\n")
    tt= random.randint(15000,20000)
    if tt not in procarr:
        return tt
    else:
        getPort()

def is_ipv4(ip):
    try:
        socket.inet_pton(socket.AF_INET, ip)
    except AttributeError:  # no inet_pton here, sorry
        try:
            socket.inet_aton(ip)
        except socket.error:
            return False
        return ip.count('.') == 3
    except socket.error:  # not a valid ip
        return False
    return True
 
 
def is_ipv6(ip):
    try:
        socket.inet_pton(socket.AF_INET6, ip)
    except socket.error:  # not a valid ip
        return False
    return True

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Send Tcp for tethering')
    parser.add_argument("-s", "--srcip", type=str, default = None)
    parser.add_argument("-d", "--destip", type=str, default = None)
    parser.add_argument("-p", '--dport', type=int, default=80)
    parser.add_argument("-w", '--window', type=int, default=8192)
    parser.add_argument("-t", '--ttl', type=int, default=64)
    parser.add_argument("-v", '--version', type=int, default=6)
    args = parser.parse_args()

    src_ip = args.srcip
    dst_ip = args.destip
    src_port = get_port()
    dst_port = args.dport
    window_size = args.window
    ttl = args.ttl
    version = args.version

    if (4 != version and 6 != version):
        print("invalid version. check the -v")
        exit(-1)

    if args.srcip == None or args.destip == None:
        print("invalid ip. check the -s or -d")
        exit(-1)

    if (6 == version):
        if is_ipv4(src_ip) or is_ipv4(dst_ip):
            print("invalid ipv6. check the -s or -d or -v")
            exit(-1)
    else:
        if is_ipv6(src_ip) or is_ipv6(dst_ip):
            print("invalid ipv4. check the -s or -d or -v")
            exit(-1)

    # data = 'GET / HTTP/1.1\r\n\r\n'
    data = r"GET / HTTP/1.1\r\nHost: [{0}]:{1}\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 12\r\nClientsenddata: Hello Server\r\nAccept-Encoding: gzip\r\n\r\nHello Server".format(dst_ip, dst_port)

    if (0 != disable_reset(src_ip, dst_ip, src_port, dst_port, version)):
        print("Disable RST package fail.")
        exit(-1)

    try:
        # SYN
        syn_seq = 100
        global syn_ip
        if (6 == version):
            syn_ip = IPv6(src=src_ip, dst=dst_ip)
        else:
            syn_ip = IP(src=src_ip, dst=dst_ip)

        syn_ip.ttl = ttl

        syn_tcp = TCP(sport=src_port, dport=dst_port, seq=syn_seq, flags="S", window=window_size)
        ans = sr1(syn_ip/syn_tcp ,verbose=False)

        if ("SA" != ans.flags):
            print("Recv the second package fail, the flag not SA. ans.flags={0}".format(ans.flags))
            exit(-1)

        # ACK
        seq = syn_seq + 1
        ack = ans.seq + 1
        ack_tcp = TCP(sport=src_port, dport=dst_port, seq=seq, ack=ack, flags="A")
        send(syn_ip/ack_tcp, verbose=False)

        # Data
        data_tcp = TCP(sport=src_port, dport=dst_port, seq=syn_seq + 1, ack=ack, flags=24)
        data_ans = sr1(syn_ip/data_tcp/data, verbose=False)

        # print("data_ans:data_ans.seq={0}, data_ans.ack={1}".format(data_ans.seq, data_ans.ack))

        # # Data ACK
        # data_ack_tcp = TCP(sport=src_port, dport=dst_port, seq=data_ans.ack, ack= data_ans.seq + 1 + len(data), flags='A')
        # send(syn_ip/data_ack_tcp, verbose=False)

        # FIN1
        fin1_tcp = TCP(sport=src_port, dport=dst_port, seq=data_ans.ack, ack= data_ans.seq, flags='F')
        fin_ack_ans = sr1(syn_ip/fin1_tcp, verbose=False)

        # # FIN2
        # fin2_tcp = TCP(sport=src_port, dport=dst_port, seq=fin_ack_ans.ack, ack= fin_ack_ans.seq + 1, flags='A')
        # fin_ans = sr1(syn_ip/fin2_tcp, verbose=False)
       
    except Exception,e:
        print e


    if (0 != enable_reset(src_ip, dst_ip, src_port, dst_port, version)):
        print("Enable RST package fail.")
        exit(-1)

    
    if (6 == version):
        print("Ok, IPv6 Scapy 3 Way Done.")
    else:
        print("Ok, IPv4 Scapy 3 Way Done.")

    exit(0)
