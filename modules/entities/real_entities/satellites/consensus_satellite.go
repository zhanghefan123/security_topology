package satellites

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	pbNode "zhanghefan123/security_topology/services/position/protobuf/node"
)

// ConsensusSatellite 共识卫星
type ConsensusSatellite struct {
	*normal_node.NormalNode
	OrbitId      int      // 轨道的编号
	IndexInOrbit int      // 轨道内的编号
	StartRpcPort int      // 起始的 RPC 端口
	StartP2PPort int      // 起始的 P2P 端口
	Tle          []string // tle
}

// NewConsensusSatellite 创建新的共识卫星
func NewConsensusSatellite(nodeId, orbitId, indexInOrbit int, startRpcPort int, startP2PPort int, tle []string) *ConsensusSatellite {
	// 当前的类型
	satType := types.NetworkNodeType_ConsensusSatellite
	sat := &ConsensusSatellite{
		NormalNode:   normal_node.NewNormalNode(types.NetworkNodeType_ConsensusSatellite, nodeId, fmt.Sprintf("%s-%d", satType.String(), nodeId)),
		OrbitId:      orbitId,
		IndexInOrbit: indexInOrbit,
		StartRpcPort: startRpcPort,
		StartP2PPort: startP2PPort,
		Tle:          tle,
	}
	return sat
}

// StoreToEtcd 将节点信息存储到 etcd 之中
func (consensusSatellite *ConsensusSatellite) StoreToEtcd(etcdClient *clientv3.Client) error {
	consensusPbSatellite := &pbNode.Node{
		Type:           pbNode.NodeType_NODE_TYPE_SATELLITE,
		Id:             int32(consensusSatellite.Id),
		ContainerName:  consensusSatellite.ContainerName,
		Pid:            int32(consensusSatellite.Pid),
		Tle:            consensusSatellite.Tle,
		InterfaceDelay: make([]string, 0),
	}
	satelliteInBytes := protobuf.MustMarshal(consensusPbSatellite)
	etcdSatellitesPrefix := configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.SatellitesPrefix
	satelliteKey := fmt.Sprintf("%s/%d", etcdSatellitesPrefix, consensusSatellite.Id)
	_, err := etcdClient.Put(context.Background(), satelliteKey, string(satelliteInBytes))
	if err != nil {
		return fmt.Errorf("failed to store normal satellite into etcd: %w", err)
	}
	return nil
}

func (consensusSatellite *ConsensusSatellite) GetOrbitId() int {
	return consensusSatellite.OrbitId
}

func (consensusSatellite *ConsensusSatellite) GetIndexInOrbit() int {
	return consensusSatellite.IndexInOrbit
}
