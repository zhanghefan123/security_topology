package constellation

import "time"

type SatelliteConfig struct {
	Type    int32 `mapstructure:"type"`
	P2PPort int   `mapstructure:"p2p_port"`
	RPCPort int   `mapstructure:"rpc_port"`
}

type ConstellationConfig struct {
	OrbitNumber       int             `mapstructure:"orbit_number"`
	SatellitePerOrbit int             `mapstructure:"satellite_per_orbit"`
	StartTime         string          `mapstructure:"start_time"`
	GoStartTime       time.Time       // 转换为了 time.Time 的值
	SatelliteConfig   SatelliteConfig `mapstructure:"satellite_config"`
	ISLBandwidth      int             `mapstructure:"isl_bandwidth"`
}
