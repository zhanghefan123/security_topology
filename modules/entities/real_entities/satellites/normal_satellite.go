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

type NormalSatellite struct {
	*normal_node.NormalNode
	OrbitId      int      // 轨道的编号
	IndexInOrbit int      // 轨道内的编号
	Tle          []string // TLE 位置信息
}

// NewNormalSatellite 创建普通卫星
func NewNormalSatellite(nodeId, orbitId, indexInOrbit int, tle []string) *NormalSatellite {
	// 当前的类型
	satType := types.NetworkNodeType_NormalSatellite
	// 创建普通卫星
	sat := &NormalSatellite{
		NormalNode:   normal_node.NewNormalNode(types.NetworkNodeType_NormalSatellite, nodeId, fmt.Sprintf("%s-%d", satType.String(), nodeId)),
		OrbitId:      orbitId,
		IndexInOrbit: indexInOrbit,
		Tle:          tle,
	}
	return sat
}

// StoreToEtcd 将节点信息存储到 etcd 之中
func (normalSatellite *NormalSatellite) StoreToEtcd(etcdClient *clientv3.Client) error {
	normalPbSatellite := &pbNode.Node{
		Type:           pbNode.NodeType_NODE_TYPE_SATELLITE,
		Id:             int32(normalSatellite.Id),
		ContainerName:  normalSatellite.ContainerName,
		Pid:            int32(normalSatellite.Pid),
		Tle:            normalSatellite.Tle,
		InterfaceDelay: make([]string, 0),
	}
	satelliteInBytes := protobuf.MustMarshal(normalPbSatellite)
	etcdSatellitesPrefix := configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.SatellitesPrefix
	satelliteKey := fmt.Sprintf("%s/%d", etcdSatellitesPrefix, normalSatellite.Id)
	_, err := etcdClient.Put(context.Background(), satelliteKey, string(satelliteInBytes))
	if err != nil {
		return fmt.Errorf("failed to store normal satellite into etcd: %w", err)
	}
	return nil
}

func (normalSatellite *NormalSatellite) GetOrbitId() int {
	return normalSatellite.OrbitId
}

func (normalSatellite *NormalSatellite) GetIndexInOrbit() int {
	return normalSatellite.IndexInOrbit
}
