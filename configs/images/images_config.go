package images

type ImagesConfig struct {
	NormalSatelliteImageName string `mapstructure:"normal_satellite_image_name"`
	GroundStationImageName   string `mapstructure:"ground_station_image_name"`
	EtcdServiceImageName     string `mapstructure:"etcd_service_image_name"`
	PositionServiceImageName string `mapstructure:"position_service_image_name"`
	RouterImageName          string `mapstructure:"router_image_name"`
	NormalNodeImageName      string `mapstructure:"normal_node_image_name"`
	ConsensusNodeImageName   string `mapstructure:"consensus_node_image_name"`
	ChainMakerNodeImageName  string `mapstructure:"chain_maker_node_image_name"`
	MaliciousNodeImageName   string `mapstructure:"malicious_node_image_name"`
	LirNodeImageName         string `mapstructure:"lir_node_image_name"`
	EntranceImageName        string `mapstructure:"entrance_image_name"`
}
