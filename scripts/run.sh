#!/bin/bash

cd ..

# Check Permissions
if [ $EUID -ne 0 ]
  then echo "Please run as root"
  exit
fi

# Clean-Up Namespaces
ip netns del net1 2> /dev/null
ip netns del net2 2> /dev/null
ip netns del net3 2> /dev/null

# Adding Namespaces
ip netns add net1
ip netns add net2
ip netns add net3

# Setting Network Paths (2 MP-TCP interfaces, 1 TCP)
ip link add veth1.1 type veth peer name veth1.2
ip link set veth1.1 netns net1
ip link set veth1.2 netns net2

ip link add veth2.1 type veth peer name veth2.2
ip link set veth2.1 netns net1
ip link set veth2.2 netns net2

ip link add veth3.1 type veth peer name veth3.2
ip link set veth3.1 netns net2
ip link set veth3.2 netns net3


# Configuring Namespace 1 (MP-TCP Client)
NS1="ip netns exec net1"
$NS1 ip link set lo up
$NS1 ip link set veth1.1 up
$NS1 ip addr add dev veth1.1 192.168.0.1/32
$NS1 ip link set veth2.1 up
$NS1 ip addr add dev veth2.1 192.168.0.3/32

$NS1 ip mptcp limits set add_addr_accepted 0 subflows 1
$NS1 ip mptcp endpoint add 192.168.0.3 dev veth2.1 subflow

$NS1 ip r add 192.168.0.2/32 dev veth1.1


# Configuring Namespace 2 (MP-TCP Proxy Server)
NS2="ip netns exec net2"
$NS2 ip link set lo up
$NS2 ip link set veth1.2 up
$NS2 ip addr add dev veth1.2 192.168.0.2/32
$NS2 ip link set veth2.2 up

$NS2 ip mptcp limits set add_addr_accepted 0 subflows 1

$NS2 ip link set veth3.1 up
$NS2 ip addr add dev veth3.1 10.200.0.1/24

$NS2 ip r add 192.168.0.3/32 dev veth2.2
$NS2 ip r add 192.168.0.1/32 dev veth1.2
$NS2 ip addr add dev veth3.1 10.60.0.1/32


# Configuring Namespace 3 (TCP Server)
NS3="ip netns exec net3"
$NS3 ip link set lo up
$NS3 ip link set veth3.2 up
$NS3 ip addr add dev veth3.2 10.200.0.2/24
$NS3 ip r add default dev veth3.2
sleep 1


# Compiling Binaries
cd server
go build
cd ../proxy
go build
cd ../client
go build

# Enable Packet Capture
while getopts ":d" option; do
   case $option in
      d)
      	echo "Dump Enabled!"
      	$NS1 tcpdump -i any -w ../pcaps/mptcp.pcap &
	$NS2 tcpdump -i veth3.1 -w ../pcaps/tcp.pcap &
   esac
done


# Starting the Testbed
$NS3 ../server/server &
#$NS3 nc -l 5001 &
#$NS3 iperf3 -s -B 10.200.0.2 -p 5001 &
sleep 1

$NS2 ../proxy/proxy &
sleep 1

$NS1 ../client/client
#$NS1 mptcpize run nc 192.168.0.2 8080
#$NS1 mptcpize run iperf3 -c 192.168.0.2 -p 8080 -t 10 -i 1

# Killing Proces
kill -9 $(pidof proxy)
kill -9 $(pidof server)
kill -9 $(pidof client)
