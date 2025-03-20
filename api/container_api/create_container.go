package container_api

import (
	"fmt"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/api/container_api/create_apis"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
)

// createFuncTemplate 创建函数模板
type createFuncTemplate func(client *docker.Client, abstractNode *node.AbstractNode) error

// createFuncs 创建函数
var createFuncs = map[types.NetworkNodeType]createFuncTemplate{
	types.NetworkNodeType_NormalSatellite: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateNormalSatellite(client, abstractNode.ActualNode.(*satellites.NormalSatellite))
	},
	types.NetworkNodeType_GroundStation: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateGroundStation(client, abstractNode.ActualNode.(*ground_station.GroundStation))
	},
	types.NetworkNodeType_EtcdService: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateEtcdNode(client, abstractNode.ActualNode.(*etcd.EtcdNode))
	},
	types.NetworkNodeType_PositionService: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateRealTimePositionService(client, abstractNode.ActualNode.(*position.PositionService))
	},
	types.NetworkNodeType_Router: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateRouter(client, abstractNode.ActualNode.(*nodes.Router))
	},
	types.NetworkNodeType_NormalNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateNormalNode(client, abstractNode.ActualNode.(*normal_node.NormalNode))
	},
	types.NetworkNodeType_ConsensusNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateConsensusNode(client, abstractNode.ActualNode.(*nodes.ConsensusNode))
	},
	types.NetworkNodeType_ChainMakerNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateChainMakerNode(client, abstractNode.ActualNode.(*nodes.ChainmakerNode))
	},
	types.NetworkNodeType_MaliciousNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateMaliciousNode(client, abstractNode.ActualNode.(*nodes.MaliciousNode))
	},
	types.NetworkNodeType_LirNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateLirNode(client, abstractNode.ActualNode.(*nodes.LiRNode), int(abstractNode.Node.ID())+1)
	},
	types.NetworkNodeType_Entrance: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateEntrance(client, abstractNode.ActualNode.(*nodes.Entrance))
	},
	types.NetworkNodeType_FabricPeerNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateFabricPeerNode(client, abstractNode.ActualNode.(*nodes.FabricPeerNode))
	},
	types.NetworkNodeType_FabricOrderNode: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateFabricOrderNode(client, abstractNode.ActualNode.(*nodes.FabricOrderNode))
	},
	types.NetworkNodeType_LiRSatellite: func(client *docker.Client, abstractNode *node.AbstractNode) error {
		return create_apis.CreateLiRSatellite(client, abstractNode.ActualNode.(*satellites.LiRSatellite), int(abstractNode.Node.ID())+1)
	},
}

// CreateContainer 创建 container
func CreateContainer(client *docker.Client, node *node.AbstractNode) error {
	if createFunc, ok := createFuncs[node.Type]; ok {
		return createFunc(client, node)
	} else {
		return fmt.Errorf("create container error: unknown node type %v", node.Type)
	}
}
