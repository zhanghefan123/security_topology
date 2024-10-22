package position

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type PositionService struct {
	*normal_node.NormalNode
	EtcdListenAddr         string // etcd 监听的地址
	EtcdClientPort         int    // etcd 客户端的端口
	EtcdISLsPrefix         string // 星间链路在 etcd 之中的前缀
	EtcdSatellitesPrefix   string // 卫星在 etcd 之中的前缀
	ConstellationStartTime string // 星座的启动时间
	UpdateInterval         int    // 更新的时间间隔
}

// NewPositionService 创建新的位置服务对象
func NewPositionService(status types.NetworkNodeStatus, etcdListenAddr string, etcdClientPort int,
	etcdISLsPrefix, etcdSatellitesPrefix,
	constellationStartTime string, UpdateInterval int) *PositionService {
	return &PositionService{
		NormalNode: &normal_node.NormalNode{
			Status: status,
		},
		EtcdListenAddr:         etcdListenAddr,
		EtcdClientPort:         etcdClientPort,
		EtcdISLsPrefix:         etcdISLsPrefix,
		EtcdSatellitesPrefix:   etcdSatellitesPrefix,
		ConstellationStartTime: constellationStartTime,
		UpdateInterval:         UpdateInterval,
	}
}
