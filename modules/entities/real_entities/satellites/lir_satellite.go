package satellites

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/utils/protobuf"
	pbNode "zhanghefan123/security_topology/services/update/protobuf/node"
)

type LiRSatellite struct {
	*normal_node.NormalNode
	OrbitId      int      // 轨道的编号
	IndexInOrbit int      // 轨道内的编号
	Tle          []string // TLE 位置信息
}

func NewLiRSatellite(nodeId, orbitId, indexInOrbit int, tle []string) *LiRSatellite {
	// 当前的类型
	nodeType := types.NetworkNodeType_LiRSatellite
	// 创建LiRSatellite
	lirSatellite := &LiRSatellite{
		NormalNode:   normal_node.NewNormalNode(nodeType, nodeId, fmt.Sprintf("%s-%d", nodeType.String(), nodeId)),
		OrbitId:      orbitId,
		IndexInOrbit: indexInOrbit,
		Tle:          tle,
	}
	return lirSatellite
}

func (lirSatellite *LiRSatellite) StoreToEtcd(etcdClient *clientv3.Client) error {
	// 卫星的下一个接口索引就是第一个 gsl 接口索引
	nextGslIfIdx := lirSatellite.Ifidx

	normalPbSatellite := &pbNode.Node{
		Type:           pbNode.NodeType_NODE_TYPE_SATELLITE,
		Id:             int32(lirSatellite.Id),
		ContainerName:  lirSatellite.ContainerName,
		Pid:            int32(lirSatellite.Pid),
		Tle:            lirSatellite.Tle,
		InterfaceDelay: make([]string, 0),
		IfIdx:          int32(nextGslIfIdx),
	}
	satelliteInBytes := protobuf.MustMarshal(normalPbSatellite)
	etcdSatellitesPrefix := configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.SatellitesPrefix
	satelliteKey := fmt.Sprintf("%s/%d", etcdSatellitesPrefix, lirSatellite.Id)
	_, err := etcdClient.Put(context.Background(), satelliteKey, string(satelliteInBytes))
	if err != nil {
		return fmt.Errorf("failed to store normal satellite into etcd: %w", err)
	}
	return nil
}
