package constellation

import (
	"context"
	"github.com/c-robinson/iplib/v2"
	"github.com/coreos/etcd/clientv3"
	docker "github.com/docker/docker/client"
	"time"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	constellationLogger = logger.GetLogger(logger.ModuleConstellation)
)

// Parameters 星座参数
type Parameters struct {
	OrbitNumber       int // 轨道数量
	SatellitePerOrbit int // 每轨道卫星数量
}

// SatelliteParameters 卫星参数
type SatelliteParameters struct {
	SatelliteType      types.NetworkNodeType // 卫星的类型
	SatelliteImageName string                // 卫星镜像名称
	SatelliteP2PPort   int                   // 启始 p2p 端口
	SatelliteRPCPort   int                   // 启始 rpc 端口
}

// Constellation 星座
type Constellation struct {
	*Parameters                                   // 星座基本参数
	*SatelliteParameters                          // 卫星基本参数
	client                   *docker.Client       // 用来创建、停止、开启容器的客户端
	etcdClient               *clientv3.Client     // etcd client 用于存取监听键值对
	startTime                time.Time            // 星座的启动时间
	SubNets                  []iplib.Net4         // 子网数量
	Satellites               []*node.AbstractNode // 所有卫星
	AllSatelliteLinks        []*link.AbstractLink // 所有的卫星链路
	InterOrbitSatelliteLinks []*link.AbstractLink // 所有轨间链路
	IntraOrbitSatelliteLinks []*link.AbstractLink // 所有轨内链路
	initModules              map[string]struct{}  // 初始化模块
	startModules             map[string]struct{}  // 启动模块
	serviceContext           context.Context      // 服务上下文
	serviceContextCancelFunc context.CancelFunc   // 服务上下文的取消函数
	etcdService              *node.AbstractNode   // etcd 服务
	positionService          *node.AbstractNode   // position 服务
}

// NewConstellation 创建一个新的空的星座
func NewConstellation(client *docker.Client, etcdClient *clientv3.Client, startTime time.Time) *Constellation {
	orbitNumber := configs.TopConfiguration.ConstellationConfig.OrbitNumber
	satellitePerOrbit := configs.TopConfiguration.ConstellationConfig.SatellitePerOrbit
	constellation := &Constellation{
		Parameters: &Parameters{
			OrbitNumber:       orbitNumber,
			SatellitePerOrbit: satellitePerOrbit,
		},
		SatelliteParameters: &SatelliteParameters{
			SatelliteType:      types.NetworkNodeType(configs.TopConfiguration.ConstellationConfig.SatelliteConfig.Type),
			SatelliteImageName: configs.TopConfiguration.ConstellationConfig.SatelliteConfig.ImageName,
			SatelliteRPCPort:   configs.TopConfiguration.ConstellationConfig.SatelliteConfig.RPCPort,
			SatelliteP2PPort:   configs.TopConfiguration.ConstellationConfig.SatelliteConfig.P2PPort,
		},
		client:                   client,
		etcdClient:               etcdClient,
		startTime:                startTime,
		Satellites:               make([]*node.AbstractNode, 0),
		InterOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		IntraOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		initModules:              make(map[string]struct{}),
	}
	constellationLogger.Infof("create new constellation")
	return constellation
}
