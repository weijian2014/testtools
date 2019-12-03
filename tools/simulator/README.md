# Tips:
    Need to use the OS prepared for the simulator, please contact weijian45590@bill-jc.com
    --tcp using non-split tcp mode

# Build Simulator:
    apt-get install build-essential
    CURR_DIR=`pwd`
    mkdir -p /tmp/simulator_build
    cd /tmp/simulator_build
    cmake ${CURR_DIR}
    make


# Usage: Simulator
    Options:
     -s                        String, the source IP addrss, must be specified.
     -d                        String, the destination IP addrss, must be specified.
     -l, --sport                  Int, the source port, optional, default will be automatically assigned.
     -r, --dport                  Int, the destination port, must be specified.
     -w, --win                   Bool, using windows TCP header options of the raw socket, optional, default true.
     -i, --ios                   Bool, using ios TCP header options of the raw socket, optional, default false.
     -t, --tcp                   Bool, using TCP packet of raw socket, for non-split tcp mode, optional, default false, will using HTTP packet.
     -h, --help                Print this message and exit.
    Examples:
     ./Simulator -w -s 1.0.0.1 -d 2.0.0.2 -r 8888
     ./Simulator -w -s 1.0.0.1 -d 2.0.0.2 -l 6666 -r 8888
     ./Simulator -i -s 1.0.0.1 -d 2.0.0.2 -l 6666 -r 8888 -t
     ./Simulator --ios -s 1.0.0.1 -d 2.0.0.2 --sport 6666--dport 8888 --tcp