docker rm -f $(docker ps -aq)
ip link show type veth | awk '{print $2}' | sed 's/@.*//' | xargs -n1 sudo ip link delete