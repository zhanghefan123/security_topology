package container_api

import (
	"fmt"
	docker "github.com/docker/docker/client"
	"zhanghefan123/security_topology/api/container_api/create_apis"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
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
			return fmt.Errorf("CreateNormalSatellite err: %w", err)
		}
	case types.NetworkNodeType_ConsensusSatellite:
		sat, _ := node.ActualNode.(*satellites.ConsensusSatellite)
		err = create_apis.CreateConsensusSatellite(client, sat)
		if err != nil {
			return fmt.Errorf("CreateConsensusSatellite err: %w", err)
		}
	case types.NetworkNodeType_EtcdService:
		etcdNode, _ := node.ActualNode.(*etcd.EtcdNode)
		err = create_apis.CreateEtcdNode(client, etcdNode)
		if err != nil {
			return fmt.Errorf("CreateEtcdNode err: %w", err)
		}
	case types.NetworkNodeType_PositionService:
		positionService, _ := node.ActualNode.(*position.PositionService)
		err = create_apis.CreatePositionService(client, positionService)
		if err != nil {
			return fmt.Errorf("CreatePositionService err: %w", err)
		}
	case types.NetworkNodeType_Router:
		router, _ := node.ActualNode.(*nodes.Router)
		err = create_apis.CreateRouter(client, router)
		if err != nil {
			return fmt.Errorf("CreateRouter err: %w", err)
		}
	case types.NetworkNodeType_NormalNode:
		normalNode, _ := node.ActualNode.(*normal_node.NormalNode)
		err = create_apis.CreateNormalNode(client, normalNode)
		if err != nil {
			return fmt.Errorf("CreateNormalNode err: %w", err)
		}
	case types.NetworkNodeType_ConsensusNode:
		consensusNode, _ := node.ActualNode.(*nodes.ConsensusNode)
		err = create_apis.CreateConsensusNode(client, consensusNode)
		if err != nil {
			return fmt.Errorf("CreateConsensusNode err: %w", err)
		}
	case types.NetworkNodeType_ChainMakerNode:
		chainMakerNode, _ := node.ActualNode.(*nodes.ChainmakerNode)
		err = create_apis.CreateChainMakerNode(client, chainMakerNode)
		if err != nil {
			return fmt.Errorf("CreateChainMakerNode err: %w", err)
		}
	case types.NetworkNodeType_MaliciousNode:
		maliciousNode, _ := node.ActualNode.(*nodes.MaliciousNode)
		err = create_apis.CreateMaliciousNode(client, maliciousNode)
		if err != nil {
			return fmt.Errorf("CreateMaliciousNode err: %w", err)
		}
	case types.NetworkNodeType_LirNode:
		lirNode, _ := node.ActualNode.(*nodes.LiRNode)
		err = create_apis.CreateLirNode(client, lirNode, int(node.Node.ID()))
		if err != nil {
			return fmt.Errorf("CreateLirNode err: %w", err)
		}
	}
	return nil
}
