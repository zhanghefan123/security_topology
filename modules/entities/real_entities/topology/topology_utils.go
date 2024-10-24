package topology

import "net"

// GetChainMakerNodeListenAddresses 获取长安链节点的监听地址
func (t *Topology) GetChainMakerNodeListenAddresses() []string {
	listeningAddresses := make([]string, 0)
	for _, node := range t.ChainmakerNodes {
		ip, _, _ := net.ParseCIDR(node.Interfaces[0].Ipv4Addr)
		listeningAddresses = append(listeningAddresses, ip.String())
	}
	return listeningAddresses
}
