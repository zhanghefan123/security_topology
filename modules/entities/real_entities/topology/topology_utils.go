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

type FabricPorts struct {
	OrdererAdminListenPort   int
	OrdererGeneralListenPort int
}

type FiscoBcosPorts struct {
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

func (t *Topology) GetFabricNodeContainerNames() []string {
	fabricNodeNames := make([]string, 0)
	for _, node := range t.FabricOrdererNodes {
		fabricNodeNames = append(fabricNodeNames, node.ContainerName)
	}
	return fabricNodeNames
}

func (t *Topology) GetFiscoBcosContainerNames() []string {
	fiscoBcosNames := make([]string, 0)
	for _, node := range t.FiscoBcosNodes {
		fiscoBcosNames = append(fiscoBcosNames, node.ContainerName)
	}
	return fiscoBcosNames
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

// GetContainerNameToChainMakerPortMapping 获取所有共识节点从容器名到
func (t *Topology) GetContainerNameToChainMakerPortMapping() (map[string]*ChainmakerPorts, error) {
	portMapping := make(map[string]*ChainmakerPorts)
	p2pStartPort := configs.TopConfiguration.ChainMakerConfig.P2pStartPort
	rpcStartPort := configs.TopConfiguration.ChainMakerConfig.RpcStartPort
	// 进行 chainMaker 的 port Mapping 的添加
	for _, chainMakerNode := range t.ChainmakerNodes {
		p2pPort := chainMakerNode.Id + p2pStartPort - 1
		rpcPort := chainMakerNode.Id + rpcStartPort - 1
		portMapping[chainMakerNode.ContainerName] = &ChainmakerPorts{p2pPort, rpcPort}
	}
	return portMapping, nil
}

func (t *Topology) GetContainerNameToFabricPortMapping() (map[string]*FabricPorts, error) {
	portMapping := make(map[string]*FabricPorts)
	// 进行 Fabric 的 portMapping 的添加
	for _, fabricOrder := range t.FabricOrdererNodes {
		fabricAdminListenPort := configs.TopConfiguration.FabricConfig.OrderAdminListenStartPort + fabricOrder.Id
		fabricGeneralListenPort := configs.TopConfiguration.FabricConfig.OrderGeneralListenStartPort + fabricOrder.Id
		portMapping[fabricOrder.ContainerName] = &FabricPorts{fabricAdminListenPort,
			fabricGeneralListenPort}
	}
	return portMapping, nil
}

func (t *Topology) GetContainerNameToFiscoBcosPortMapping() (map[string]*FiscoBcosPorts, error) {
	portMapping := make(map[string]*FiscoBcosPorts)
	// 进行 fisco-bcos 的 portMapping 的添加
	for _, fiscoBcosNode := range t.FiscoBcosNodes {
		p2pPort := configs.TopConfiguration.FiscoBcosConfig.P2pStartPort + fiscoBcosNode.Id - 1
		rpcPort := configs.TopConfiguration.FiscoBcosConfig.RpcStartPort + fiscoBcosNode.Id - 1
		portMapping[fiscoBcosNode.ContainerName] = &FiscoBcosPorts{p2pPort,
			rpcPort}
	}
	return portMapping, nil
}

//func (t *Topology) ChainMakerEnabled() bool {
//	return len(t.ChainmakerNodes) > 0
//}
//
//func (t *Topology) FabricEnabled() bool {
//	return len(t.FabricOrdererNodes) > 0 && len(t.FabricPeerNodes) > 0
//}
//
//func (t *Topology) FiscoBcosEnabled() bool {
//	return len(t.FiscoBcosNodes) > 0
//}
