package constellation

import "time"

type SatelliteConfig struct {
	Type    int32 `mapstructure:"type"`
	P2PPort int   `mapstructure:"p2p_port"`
	RPCPort int   `mapstructure:"rpc_port"`
}

type ConstellationConfig struct {
	// 这两个参数, 前端会进行传入, 所以不用进行单独的在 yml 文件之中的配置
	// OrbitNumber       int             `mapstructure:"orbit_number"`
	// SatellitePerOrbit int             `mapstructure:"satellite_per_orbit"`
	StartTime       string          `mapstructure:"start_time"`
	GoStartTime     time.Time       // 转换为了 time.Time 的值
	SatelliteConfig SatelliteConfig `mapstructure:"satellite_config"`
	ISLBandwidth    int             `mapstructure:"isl_bandwidth"`
	GSLBandwidth    int             `mapstructure:"gsl_bandwidth"`
}
