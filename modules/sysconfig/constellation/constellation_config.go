package constellation

type SatelliteConfig struct {
	Type      int32  `mapstructure:"type"`
	ImageName string `mapstructure:"image_name"`
	P2PPort   int    `mapstructure:"p2p_port"`
	RPCPort   int    `mapstructure:"rpc_port"`
}

type ConstellationConfig struct {
	OrbitNumber       int             `mapstructure:"orbit_number"`
	SatellitePerOrbit int             `mapstructure:"satellite_per_orbit"`
	SatelliteConfig   SatelliteConfig `mapstructure:"satellite_config"`
}
