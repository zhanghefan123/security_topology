package topology

import (
	"fmt"
	"net"
)

// GetChainMakerNodeListenAddresses 获取长安链节点的监听地址
func (t *Topology) GetChainMakerNodeListenAddresses() []string {
	listeningAddresses := make([]string, 0)
	for _, node := range t.ChainmakerNodes {
		ip, _, _ := net.ParseCIDR(node.Interfaces[0].Ipv4Addr)
		listeningAddresses = append(listeningAddresses, ip.String())
	}
	return listeningAddresses
}

// GetContainerNameToAddressMapping 获取所有节点的从容器名称到地址的一个映射
func (t *Topology) GetContainerNameToAddressMapping() (map[string]string, error) {
	addressMapping := make(map[string]string)
	for _, abstractNode := range t.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToAddressMapping abstract node error: %w", err)
		}
		ip, _, _ := net.ParseCIDR(normalNode.Interfaces[0].Ipv4Addr)
		addressMapping[normalNode.ContainerName] = ip.String()
	}
	return addressMapping, nil
}
