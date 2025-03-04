package constellation

import (
	"fmt"
	"net"
)

// GetContainerNameToAddressMapping 获取容器名到地址的映射
func (c *Constellation) GetContainerNameToAddressMapping() (map[string][]string, error) {
	addressMapping := make(map[string][]string)
	allAbstractNodes := append(c.SatelliteAbstractNodes, c.GroundStationAbstractNodes...)
	for _, abstractNode := range allAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToAddressMapping Error: %v", err)
		}
		ip, _, _ := net.ParseCIDR(normalNode.Interfaces[0].SourceIpv4Addr)
		ipv6, _, _ := net.ParseCIDR(normalNode.Interfaces[0].SourceIpv6Addr)
		addressMapping[normalNode.ContainerName] = []string{ip.String(), ipv6.String()}
	}
	return addressMapping, nil
}

// GetContainerNameToGraphIdMapping 获取容器名到 id 的映射
func (c *Constellation) GetContainerNameToGraphIdMapping() (map[string]int, error) {
	idMapping := make(map[string]int)
	allAbstractNodes := append(c.SatelliteAbstractNodes, c.GroundStationAbstractNodes...)
	for _, abstractNode := range allAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return nil, fmt.Errorf("GetContainerNameToGraphIdMapping Error: %v", err)
		}
		idMapping[normalNode.ContainerName] = int(abstractNode.Node.ID() + 1)
	}
	return idMapping, nil
}
