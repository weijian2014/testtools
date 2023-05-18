# Tips:
    TCP raw socket simulate to IOS SYN

# Build Simulator:
    apt-get install build-essential
    CURR_DIR=`pwd`
    mkdir -p /tmp/simulator_build
    cd /tmp/simulator_build
    cmake ${CURR_DIR}
    make

# Usage: Simulator
   Options:
   -h, --help                Print this message and exit.
   -p, --debug               Print debug log, optional, default false.
   -v, --verify              Verify the packets received from the server, optional, default false.
   -t, --ttl                    Int, the TTL of SYNC packet, optional, default 64.
   -s                        String, the source IP addrss, must be specified.
   -d                        String, the destination IP addrss, must be specified.
   -l, --sport                  Int, the source port, optional, default will be automatically assigned.
   -r, --dport                  Int, the destination port, must be specified.
   Examples:
   ./DpiIosSimulator -s 1.0.0.1 -d 2.0.0.2 -l 6666 -r 8888 -t
   ./DpiIosSimulator -s 1.0.0.1 -d 2.0.0.2 --sport 6666--dport 8888 --tcp
