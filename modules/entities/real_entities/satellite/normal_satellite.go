package satellite

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type NormalSatellite struct {
	*normal_node.NormalNode
	OrbitId      int    // 轨道的编号
	IndexInOrbit int    // 轨道内的编号
	ImageName    string // 镜像名称
}

// NewNormalSatellite 创建普通卫星
func NewNormalSatellite(nodeId, orbitId, indexInOrbit int, imageName string) *node.AbstractNode {
	satType := types.NetworkNodeType_NormalSatellite
	sat := &NormalSatellite{
		NormalNode: normal_node.NewNormalNode(types.NetworkNodeStatus_Logic, nodeId, 1,
			fmt.Sprintf("%s-%d", satType.String(), nodeId)),
		OrbitId:      orbitId,
		IndexInOrbit: indexInOrbit,
		ImageName:    imageName,
	}
	return &node.AbstractNode{
		Type:       satType,
		ActualNode: sat,
	}
}
