package satellite

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// ConsensusSatellite 共识卫星
type ConsensusSatellite struct {
	*normal_node.NormalNode
	OrbitId      int    // 轨道的编号
	IndexInOrbit int    // 轨道内的编号
	ImageName    string // 镜像名称
	StartRpcPort int    // 起始的 RPC 端口
	StartP2PPort int    // 起始的 P2P 端口
}

// NewConsensusSatellite 创建新的共识卫星
func NewConsensusSatellite(nodeId, orbitId, indexInOrbit int, imageName string, startRpcPort int, startP2PPort int) *node.AbstractNode {
	satType := types.NetworkNodeType_ConsensusSatellite
	sat := &ConsensusSatellite{
		NormalNode: normal_node.NewNormalNode(types.NetworkNodeStatus_Logic, nodeId, 1,
			fmt.Sprintf("%s-%d", satType.String(), nodeId)),
		OrbitId:      orbitId,
		IndexInOrbit: indexInOrbit,
		ImageName:    imageName,
		StartRpcPort: startRpcPort,
		StartP2PPort: startP2PPort,
	}
	return &node.AbstractNode{
		Type:       satType,
		ActualNode: sat,
	}
}
