package topology

import (
	"fmt"
	"net"
	"zhanghefan123/security_topology/configs"
)

type ChainmakerPorts struct {
	p2pPort int
	rpcPort int
}

// GetChainMakerNodeListenAddresses 获取长安链节点的监听地址
func (t *Topology) GetChainMakerNodeListenAddresses() []string {
	listeningAddresses := make([]string, 0)
	for _, node := range t.ChainmakerNodes {
		ip, _, _ := net.ParseCIDR(node.Interfaces[0].SourceIpv4Addr)
		listeningAddresses = append(listeningAddresses, ip.String())
	}
	return listeningAddresses
}

// GetChainMakerNodeContainerNames 获取所有长安链容器的名称
func (t *Topology) GetChainMakerNodeContainerNames() []string {
	chainMakerNodeNames := make([]string, 0)
	for _, node := range t.ChainmakerNodes {
		chainMakerNodeNames = append(chainMakerNodeNames, node.ContainerName)
	}
	return chainMakerNodeNames
}

// GetContainerNameToAddressMapping 获取所有节点的从容器名称到地址的一个映射 (修改成了从容器名称到 ipv4 和 ipv6 地址的映射)
func (t *Topology) GetContainerNameToAddressMapping() (map[string][]string, error) {
	addressMapping := make(map[string][]string)
	for _, abstractNode := range t.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToAddressMapping abstract node error: %w", err)
		}
		ip, _, _ := net.ParseCIDR(normalNode.Interfaces[0].SourceIpv4Addr)
		ipv6, _, _ := net.ParseCIDR(normalNode.Interfaces[0].SourceIpv6Addr)
		addressMapping[normalNode.ContainerName] = []string{ip.String(), ipv6.String()}
	}
	return addressMapping, nil
}

// GetContainerNameToGraphIdMapping 获取从所有节点的容器名称到图节点id的一个映射
func (t *Topology) GetContainerNameToGraphIdMapping() (map[string]int, error) {
	idMapping := make(map[string]int)
	for _, abstractNode := range t.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToGraphIdMapping abstract node error: %w", err)
		}
		idMapping[normalNode.ContainerName] = int(abstractNode.Node.ID() + 1)
	}
	return idMapping, nil
}

// GetContainerNameToPortMapping 获取所有共识节点从容器名到
func (t *Topology) GetContainerNameToPortMapping() (map[string]*ChainmakerPorts, error) {
	portMapping := make(map[string]*ChainmakerPorts)
	p2pStartPort := configs.TopConfiguration.ChainMakerConfig.P2pStartPort
	rpcStartPort := configs.TopConfiguration.ChainMakerConfig.RpcStartPort
	for _, chainMakerNode := range t.ChainmakerNodes {
		p2pPort := chainMakerNode.Id + p2pStartPort - 1
		rpcPort := chainMakerNode.Id + rpcStartPort - 1
		portMapping[chainMakerNode.ContainerName] = &ChainmakerPorts{p2pPort, rpcPort}
	}
	return portMapping, nil
}
