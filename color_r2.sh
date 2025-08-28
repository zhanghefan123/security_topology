# R2 的配置文件
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
# hostname r2
# no ipv6 forwarding
# !
#    mpls lsp 100 10.134.0.1 implicit-null
# !
# --------------------------------------------

# 地址配置
# --------------------------------------------
#enp3s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
#        inet 10.0.2.3  netmask 255.255.255.0  broadcast 10.0.2.255
#        ether a8:b8:e0:09:c7:66  txqueuelen 1000  (Ethernet)
#        RX packets 3498  bytes 698506 (698.5 KB)
#        RX errors 0  dropped 109  overruns 0  frame 0
#        TX packets 2942  bytes 454743 (454.7 KB)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80800000-808fffff
#
#enp4s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
#        inet 10.134.0.10  netmask 255.255.255.0  broadcast 10.134.0.255
#        ether a8:b8:e0:09:c7:67  txqueuelen 1000  (Ethernet)
#        RX packets 836  bytes 134146 (134.1 KB)
#        RX errors 0  dropped 141  overruns 0  frame 0
#        TX packets 2148  bytes 362422 (362.4 KB)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80600000-806fffff
#
#enp5s0: flags=4099<UP,BROADCAST,MULTICAST>  mtu 1500
#        ether a8:b8:e0:09:c7:68  txqueuelen 1000  (Ethernet)
#        RX packets 0  bytes 0 (0.0 B)
#        RX errors 0  dropped 0  overruns 0  frame 0
#        TX packets 0  bytes 0 (0.0 B)
#        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
#        device memory 0x80400000-804fffff
# --------------------------------------------


# Let NetworkManager manage all devices on this system
#network:
#  version: 2
#  renderer: networkd
#  ethernets:
#    enp3s0:
#      dhcp4: no
#      addresses: [10.0.2.3/24]
#    enp4s0:
#      dhcp4: no
#      addresses: [10.134.0.10/24]


sudo systemctl restart frr