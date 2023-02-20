# Tips:
    t, ttl not support
    v, version as 4(ipv4) not support

# Tcp3Way.py:
    python Tcp3Way.py -s fd14:f60:d4db:7f9:24:0:0:1 -d fd14:f60:d4db:7f9:24:0:0:1 -p 1001 -w 4096

# Usage: Tcp3Way.py
    usage: Tcp3Way.py [-h] [-s SRCIP] [-d DESTIP] [-p DPORT] [-w WINDOW] [-t TTL]
                    [-v VERSION]

    Send Tcp for tethering

    optional arguments:
    -h, --help            show this help message and exit
    -s SRCIP, --srcip SRCIP
    -d DESTIP, --destip DESTIP
    -p DPORT, --dport DPORT
    -w WINDOW, --window WINDOW
    -t TTL, --ttl TTL
    -v VERSION, --version VERSION
