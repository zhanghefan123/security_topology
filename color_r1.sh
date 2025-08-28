# R1 的配置文件
sudo modprobe mpls_router
sudo modprobe mpls_gso
sudo modprobe mpls_iptunnel
sudo sysctl -w net.ipv4.ip_forward=1
sudo sysctl -w net.ipv4.conf.all.forwarding=1
sudo sysctl -w net.mpls.platform_labels=1048575
sudo sysctl -w net.ipv6.seg6_flowlabel=1
sudo sysctl -w net.mpls.conf.enp3s0.input=1
sudo sysctl -w net.mpls.conf.enp4s0.input=1
sudo sysctl -w net.mpls.conf.enp5s0.input=1

# frr 配置文件
# --------------------------------------------
# log syslog informational
# frr version 8.1_git
# frr defaults traditional
# hostname r1
# no ipv6 forwarding
# !
#   ip route 10.134.0.1/24 10.0.2.3 label 100
# !
# --------------------------------------------

# 地址配置
# --------------------------------------------
#enp3s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
#        inet 10.0.2.8  netmask 255.255.255.0  broadcast 10.0.2.255
#        ether a8:b8:e0:09:c7:f2  txqueuelen 1000  (以太网)
#        RX packets 4840  bytes 860601 (860.6 KB)
#        RX errors 0  dropped 112  overruns 0  frame 0
#        TX packets 2261  bytes 339152 (339.1 KB)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80800000-808fffff
#
#enp4s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
#        inet 10.0.3.8  netmask 255.255.255.0  broadcast 10.0.3.255
#        inet6 fe80::8f:91a3:eaff:99c8  prefixlen 64  scopeid 0x20<link>
#        ether a8:b8:e0:09:c7:f3  txqueuelen 1000  (以太网)
#        RX packets 4018  bytes 768623 (768.6 KB)
#        RX errors 0  dropped 81  overruns 0  frame 0
#        TX packets 1669  bytes 266611 (266.6 KB)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80600000-806fffff
#
#enp5s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
#        inet 10.0.1.8  netmask 255.255.255.0  broadcast 10.0.1.255
#        inet6 fe80::916b:355b:6c85:b70c  prefixlen 64  scopeid 0x20<link>
#        ether a8:b8:e0:09:c7:f4  txqueuelen 1000  (以太网)
#        RX packets 0  bytes 0 (0.0 B)
#        RX errors 0  dropped 0  overruns 0  frame 0
#        TX packets 1878  bytes 309686 (309.6 KB)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80400000-804fffff
# --------------------------------------------

# 进行网络的配置
# # Let NetworkManager manage all devices on this system
# network:
#   version: 2
#   renderer: networkd
#   ethernets:
#     enp3s0:
#       dhcp4: no
#       addresses: [10.0.2.8/24]
#     enp5s0:
#       dhcp4: no
#       addresses: [10.0.3.8/24]
#     enp4s0:
#       dhcp4: no
#       addresses: [10.0.1.8/24]

sudo systemctl restart frr