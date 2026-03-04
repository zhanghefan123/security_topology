package container_api

import (
	"fmt"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/api/container_api/create_apis"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
)

// createFuncTemplate 创建函数模板
type createFuncTemplate func(client *docker.Client, param *ContainerCreateParameter) error

// createFuncs 创建函数
var createFuncs = map[types.NetworkNodeType]createFuncTemplate{
	types.NetworkNodeType_NormalSatellite: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateNormalSatellite(client, param.AbstractNode.ActualNode.(*satellites.NormalSatellite))
	},
	types.NetworkNodeType_GroundStation: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateGroundStation(client, param.AbstractNode.ActualNode.(*ground_station.GroundStation))
	},
	types.NetworkNodeType_EtcdService: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateEtcdNode(client, param.AbstractNode.ActualNode.(*etcd.EtcdNode))
	},
	types.NetworkNodeType_PositionService: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateRealTimePositionService(client, param.AbstractNode.ActualNode.(*position.PositionService))
	},
	types.NetworkNodeType_Router: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateRouter(client, param.AbstractNode.ActualNode.(*nodes.Router))
	},
	types.NetworkNodeType_NormalNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateNormalNode(client, param.AbstractNode.ActualNode.(*normal_node.NormalNode))
	},
	types.NetworkNodeType_ConsensusNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateConsensusNode(client, param.AbstractNode.ActualNode.(*nodes.ConsensusNode))
	},
	types.NetworkNodeType_ChainMakerNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateChainMakerNode(client, param.AbstractNode.ActualNode.(*nodes.ChainmakerNode), int(param.AbstractNode.Node.ID())+1)
	},
	types.NetworkNodeType_MaliciousNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateMaliciousNode(client, param.AbstractNode.ActualNode.(*nodes.MaliciousNode))
	},
	types.NetworkNodeType_FiscoBcosNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateFiscoBcosNode(client, param.AbstractNode.ActualNode.(*nodes.FiscoBcosNode), int(param.AbstractNode.Node.ID())+1)
	},
	types.NetworkNodeType_LirNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateLirNode(client, param.AbstractNode.ActualNode.(*nodes.LiRNode), int(param.AbstractNode.Node.ID())+1, param.NumberOfNodes)
	},
	types.NetworkNodeType_Entrance: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateEntrance(client, param.AbstractNode.ActualNode.(*nodes.Entrance))
	},
	types.NetworkNodeType_FabricPeerNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateFabricPeerNode(client, param.AbstractNode.ActualNode.(*nodes.FabricPeerNode), int(param.AbstractNode.Node.ID())+1)
	},
	types.NetworkNodeType_FabricOrderNode: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateFabricOrderNode(client, param.AbstractNode.ActualNode.(*nodes.FabricOrderNode), int(param.AbstractNode.Node.ID())+1)
	},
	types.NetworkNodeType_LiRSatellite: func(client *docker.Client, param *ContainerCreateParameter) error {
		return create_apis.CreateLiRSatellite(client, param.AbstractNode.ActualNode.(*satellites.LiRSatellite), int(param.AbstractNode.Node.ID())+1)
	},
}

// CreateContainer 创建 container
func CreateContainer(client *docker.Client, param *ContainerCreateParameter) error {
	if createFunc, ok := createFuncs[param.AbstractNode.Type]; ok {
		return createFunc(client, param)
	} else {
		return fmt.Errorf("create container error: unknown node type %v", param.AbstractNode.Type)
	}
}
