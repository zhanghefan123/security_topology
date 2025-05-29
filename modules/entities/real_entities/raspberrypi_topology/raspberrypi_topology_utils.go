package raspberrypi_topology

import (
	"fmt"
	"net"
)

// GetContainerNameToAddressMapping 获取所有节点的从容器名称到地址的一个映射 (修改成了从容器名称到 ipv4 和 ipv6 地址的映射)
func (rpt *RaspberrypiTopology) GetContainerNameToAddressMapping() (map[string][]string, error) {
	addressMapping := make(map[string][]string)
	for _, abstractNode := range rpt.AllAbstractNodes {
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
func (rpt *RaspberrypiTopology) GetContainerNameToGraphIdMapping() (map[string]int, error) {
	idMapping := make(map[string]int)
	for _, abstractNode := range rpt.AllAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToGraphIdMapping abstract node error: %w", err)
		}
		idMapping[normalNode.ContainerName] = int(abstractNode.Node.ID() + 1)
	}
	return idMapping, nil
}
