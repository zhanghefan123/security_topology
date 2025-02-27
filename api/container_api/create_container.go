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

// CreateContainer 创建 container
func CreateContainer(client *docker.Client, node *node.AbstractNode) error {
	var err error = nil
	switch node.Type {
	case types.NetworkNodeType_NormalSatellite:
		sat, _ := node.ActualNode.(*satellites.NormalSatellite)
		err = create_apis.CreateNormalSatellite(client, sat)
		if err != nil {
			return fmt.Errorf("createNormalSatellite err: %w", err)
		}
	case types.NetworkNodeType_GroundStation:
		gnd, _ := node.ActualNode.(*ground_station.GroundStation)
		err = create_apis.CreateGroundStation(client, gnd)
		if err != nil {
			return fmt.Errorf("createGroundStation err: %w", err)
		}
	case types.NetworkNodeType_EtcdService:
		etcdNode, _ := node.ActualNode.(*etcd.EtcdNode)
		err = create_apis.CreateEtcdNode(client, etcdNode)
		if err != nil {
			return fmt.Errorf("createEtcdNode err: %w", err)
		}
	case types.NetworkNodeType_PositionService:
		positionService, _ := node.ActualNode.(*position.PositionService)
		err = create_apis.CreateRealTimePositionService(client, positionService)
		if err != nil {
			return fmt.Errorf("createPositionService err: %w", err)
		}
	case types.NetworkNodeType_Router:
		router, _ := node.ActualNode.(*nodes.Router)
		err = create_apis.CreateRouter(client, router)
		if err != nil {
			return fmt.Errorf("createRouter err: %w", err)
		}
	case types.NetworkNodeType_NormalNode:
		normalNode, _ := node.ActualNode.(*normal_node.NormalNode)
		err = create_apis.CreateNormalNode(client, normalNode)
		if err != nil {
			return fmt.Errorf("createNormalNode err: %w", err)
		}
	case types.NetworkNodeType_ConsensusNode:
		consensusNode, _ := node.ActualNode.(*nodes.ConsensusNode)
		err = create_apis.CreateConsensusNode(client, consensusNode)
		if err != nil {
			return fmt.Errorf("createConsensusNode err: %w", err)
		}
	case types.NetworkNodeType_ChainMakerNode:
		chainMakerNode, _ := node.ActualNode.(*nodes.ChainmakerNode)
		err = create_apis.CreateChainMakerNode(client, chainMakerNode)
		if err != nil {
			return fmt.Errorf("createChainMakerNode err: %w", err)
		}
	case types.NetworkNodeType_MaliciousNode:
		maliciousNode, _ := node.ActualNode.(*nodes.MaliciousNode)
		err = create_apis.CreateMaliciousNode(client, maliciousNode)
		if err != nil {
			return fmt.Errorf("createMaliciousNode err: %w", err)
		}
	case types.NetworkNodeType_LirNode:
		lirNode, _ := node.ActualNode.(*nodes.LiRNode)
		err = create_apis.CreateLirNode(client, lirNode, int(node.Node.ID())+1)
		if err != nil {
			return fmt.Errorf("createLirNode err: %w", err)
		}
	case types.NetworkNodeType_Entrance:
		entrance, _ := node.ActualNode.(*nodes.Entrance)
		err = create_apis.CreateEntrance(client, entrance)
		if err != nil {
			return fmt.Errorf("create Entrance err: %w", err)
		}
	case types.NetworkNodeType_FabricPeerNode:
		fabricPeer, _ := node.ActualNode.(*nodes.FabricPeerNode)
		err = create_apis.CreateFabricPeerNode(client, fabricPeer)
		if err != nil {
			return fmt.Errorf("create fabricPeer err: %w", err)
		}
	case types.NetworkNodeType_FabricOrderNode:
		fabricOrder, _ := node.ActualNode.(*nodes.FabricOrderNode)
		err = create_apis.CreateFabricOrderNode(client, fabricOrder)
		if err != nil {
			return fmt.Errorf("create fabricOrder err: %w", err)
		}
	}
	return nil
}
