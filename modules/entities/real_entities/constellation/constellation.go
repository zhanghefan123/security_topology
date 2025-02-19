package constellation

import (
	"context"
	"github.com/c-robinson/iplib/v2"
	docker "github.com/docker/docker/client"
	"go.etcd.io/etcd/client/v3"
	"gonum.org/v1/gonum/graph/simple"
	"time"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/ground_station"
	"zhanghefan123/security_topology/modules/entities/real_entities/position_info"
	"zhanghefan123/security_topology/modules/entities/real_entities/satellites"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/position"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	ConstellationInstance *Constellation
	constellationLogger   = logger.GetLogger(logger.ModuleConstellation)
)

// Parameters -星座参数
type Parameters struct {
	OrbitNumber          int                       `json:"orbit_number"`        // 轨道数量
	SatellitePerOrbit    int                       `json:"satellite_per_orbit"` // 每轨道卫星数量
	GroundStationsParams []GroundStationParameters `json:"ground_stations"`     // 选择的地面站
}

// GroundStationParameters - 地面站参数
type GroundStationParameters struct {
	Name      string  `json:"name"`      // 地面站的名称
	Longitude float32 `json:"longitude"` // 经度
	Latitude  float32 `json:"latitude"`  // 纬度
}

// SatelliteParameters 卫星参数
type SatelliteParameters struct {
	SatelliteType    types.NetworkNodeType // 卫星的类型
	SatelliteP2PPort int                   // 启始 p2p 端口
	SatelliteRPCPort int                   // 启始 rpc 端口
}

// Constellation 星座
type Constellation struct {
	*Parameters                           // 星座基本参数
	*SatelliteParameters                  // 卫星基本参数
	client               *docker.Client   // 用来创建、停止、开启容器的客户端
	etcdClient           *clientv3.Client // etcd client 用于存取监听键值对
	startTime            time.Time        // 星座的启动时间
	Ipv4SubNets          []iplib.Net4     // ipv4 子网
	Ipv6SubNets          []iplib.Net6     // ipv6 子网

	GroundStations      []*ground_station.GroundStation  // 所有的地面站
	NormalSatellites    []*satellites.NormalSatellite    // 所有的普通卫星
	ConsensusSatellites []*satellites.ConsensusSatellite // 所有的共识卫星

	SatelliteAbstractNodes     []*node.AbstractNode // 卫星对应的抽象节点
	GroundStationAbstractNodes []*node.AbstractNode // 地面对应的抽象节点

	ContainerNameToPosition map[string]*position_info.Position // 所有节点的位置
	ConstellationGraph      *simple.DirectedGraph              // 有向图

	AllGroundSatelliteLinks    []*link.AbstractLink                     // 所有的星地链路
	AllSatelliteLinks          []*link.AbstractLink                     // 所有的卫星链路
	AllGroundSatelliteLinksMap map[string]*link.AbstractLink            // 创建星地链路映射
	AllSatelliteLinksMap       map[string]map[string]*link.AbstractLink // 创建链路映射
	InterOrbitSatelliteLinks   []*link.AbstractLink                     // 所有轨间链路
	IntraOrbitSatelliteLinks   []*link.AbstractLink                     // 所有轨内链路

	systemInitSteps  map[string]struct{} // 系统初始化步骤
	systemStartSteps map[string]struct{} // 系统启动步骤
	systemStopSteps  map[string]struct{} // 系统停止步骤

	serviceContext           context.Context           // 服务上下文
	serviceContextCancelFunc context.CancelFunc        // 服务上下文的取消函数
	etcdService              *etcd.EtcdNode            // etcd 服务
	abstractEtcdService      *node.AbstractNode        // 抽象 etcd 节点
	positionService          *position.PositionService // position 服务
	abstractPositionService  *node.AbstractNode        // 抽象位置节点

	NetworkInterfaces int // 网络接口的数量 -> 用来表征链路标识
}

// NewConstellation 创建一个新的星座实例
func NewConstellation(client *docker.Client, etcdClient *clientv3.Client, startTime time.Time, constellationParameters *Parameters) *Constellation {
	constellation := &Constellation{
		Parameters: constellationParameters,
		SatelliteParameters: &SatelliteParameters{
			SatelliteType:    types.NetworkNodeType(configs.TopConfiguration.ConstellationConfig.SatelliteConfig.Type),
			SatelliteRPCPort: configs.TopConfiguration.ConstellationConfig.SatelliteConfig.RPCPort,
			SatelliteP2PPort: configs.TopConfiguration.ConstellationConfig.SatelliteConfig.P2PPort,
		},
		client:                  client,
		etcdClient:              etcdClient,
		startTime:               startTime,
		NormalSatellites:        make([]*satellites.NormalSatellite, 0),
		ConsensusSatellites:     make([]*satellites.ConsensusSatellite, 0),
		SatelliteAbstractNodes:  make([]*node.AbstractNode, 0),
		ContainerNameToPosition: make(map[string]*position_info.Position),
		ConstellationGraph:      simple.NewDirectedGraph(),

		InterOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		IntraOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		systemInitSteps:          make(map[string]struct{}),
		systemStartSteps:         make(map[string]struct{}),
		systemStopSteps:          make(map[string]struct{}),

		AllSatelliteLinks:          make([]*link.AbstractLink, 0),
		AllGroundSatelliteLinks:    make([]*link.AbstractLink, 0),
		AllSatelliteLinksMap:       make(map[string]map[string]*link.AbstractLink),
		AllGroundSatelliteLinksMap: make(map[string]*link.AbstractLink),

		NetworkInterfaces: 0,
	}
	constellationLogger.Infof("create new constellation")
	return constellation
}
