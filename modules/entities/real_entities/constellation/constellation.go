package constellation

import (
	"github.com/c-robinson/iplib/v2"
	"zhanghefan123/security_topology/modules/config/system"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/link"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	ConstellationInstance     *Constellation
	moduleConstellationLogger = logger.GetLogger(logger.ModuleConstellation)
)

// ConstellationParameters 星座参数
type ConstellationParameters struct {
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
	*ConstellationParameters                      // 星座基本参数
	*SatelliteParameters                          // 卫星基本参数
	SubNets                  []iplib.Net4         // 子网数量
	Satellites               []*node.AbstractNode // 所有卫星
	AllSatelliteLinks        []*link.AbstractLink // 所有的卫星链路
	InterOrbitSatelliteLinks []*link.AbstractLink // 所有轨间链路
	IntraOrbitSatelliteLinks []*link.AbstractLink // 所有轨内链路
	initModules              map[string]struct{}  // 初始化模块
	startModules             map[string]struct{}  // 启动模块
}

// NewConstellation 创建一个新的空的星座
func NewConstellation() *Constellation {
	orbitNumber := system.TopConfiguration.ConstellationConfig.OrbitNumber
	satellitePerOrbit := system.TopConfiguration.ConstellationConfig.SatellitePerOrbit
	constellation := &Constellation{
		ConstellationParameters: &ConstellationParameters{
			OrbitNumber:       orbitNumber,
			SatellitePerOrbit: satellitePerOrbit,
		},
		SatelliteParameters: &SatelliteParameters{
			SatelliteType:      types.NetworkNodeType(system.TopConfiguration.ConstellationConfig.SatelliteConfig.Type),
			SatelliteImageName: system.TopConfiguration.ConstellationConfig.SatelliteConfig.ImageName,
			SatelliteRPCPort:   system.TopConfiguration.ConstellationConfig.SatelliteConfig.RPCPort,
			SatelliteP2PPort:   system.TopConfiguration.ConstellationConfig.SatelliteConfig.P2PPort,
		},
		Satellites:               make([]*node.AbstractNode, 0),
		InterOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		IntraOrbitSatelliteLinks: make([]*link.AbstractLink, 0),
		initModules:              make(map[string]struct{}),
	}
	moduleConstellationLogger.Infof("create new constellation")
	return constellation
}
