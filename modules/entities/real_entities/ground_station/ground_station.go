package ground_station

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/types"
	pbNode "zhanghefan123/security_topology/services/update/protobuf/node"
	"zhanghefan123/security_topology/utils/protobuf"
)

// GroundStation 地面站
type GroundStation struct {
	*normal_node.NormalNode
	Longitude          float32                     //  经度信息
	Latitude           float32                     // 纬度信息
	ConnectedSatellite *satellites.NormalSatellite // 之前连接的是哪一颗卫星
	RealName           string                      // 真实的地面站名称
}

// NewGroundStation 创建新的地面站实例
func NewGroundStation(nodeId int, longitude, latitude float32, realName string) *GroundStation {
	// 当前的类型
	nodeType := types.NetworkNodeType_GroundStation
	// 创建地面站
	groundStation := &GroundStation{
		NormalNode:         normal_node.NewNormalNode(nodeType, nodeId, fmt.Sprintf("%s-%d", nodeType.String(), nodeId)),
		Longitude:          longitude,
		Latitude:           latitude,
		ConnectedSatellite: nil,
		RealName:           realName,
	}
	// 将结果进行返回
	return groundStation
}

// StoreToEtcd 将地面站信息存储到 etcd 之中
func (groundStation *GroundStation) StoreToEtcd(etcdClient *clientv3.Client) error {
	normalPbGroundStation := &pbNode.Node{
		Type:           pbNode.NodeType_NODE_TYPE_GROUND_STATION,
		Id:             int32(groundStation.Id),
		ContainerName:  groundStation.ContainerName,
		Pid:            int32(groundStation.Pid),
		InterfaceDelay: make([]string, 0),
		Longitude:      groundStation.Longitude,
		Latitude:       groundStation.Latitude,
	}
	groundStationInBytes := protobuf.MustMarshal(normalPbGroundStation)
	etcdGroundStationPrefix := configs.TopConfiguration.ServicesConfig.EtcdConfig.EtcdPrefix.GroundStationsPrefix
	groundStationKey := fmt.Sprintf("%s/%d", etcdGroundStationPrefix, groundStation.Id)
	_, err := etcdClient.Put(context.Background(), groundStationKey, string(groundStationInBytes))
	if err != nil {
		return fmt.Errorf("failed to store ground station into etcd: %w", err)
	}
	return nil
}
