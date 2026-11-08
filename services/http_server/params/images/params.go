package images

type Params struct {
	NormalSatelliteImageName    string `json:"normal_satellite_image_name"`
	ConsensusSatelliteImageName string `json:"consensus_satellite_image_name"`
	GroundStationImageName      string `json:"ground_station_image_name"`
	EtcdServiceImageName        string `json:"etcd_service_image_name"`
	PositionServiceImageName    string `json:"position_service_image_name"`
	RouterImageName             string `json:"router_image_name"`
	NormalNodeImageName         string `json:"normal_node_image_name"`
	ConsensusNodeImageName      string `json:"consensus_node_image_name"`
	ChainMakerNodeImageName     string `json:"chain_maker_node_image_name"`
	MaliciousNodeImageName      string `json:"malicious_node_image_name"`
	LirNodeImageName            string `json:"lir_node_image_name"`
	EntranceImageName           string `json:"entrance_image_name"`
	FabricPeerImageName         string `json:"fabric_peer_image_name"`
	FabricOrderImageName        string `json:"fabric_order_image_name"`
	LiRSatelliteImageName       string `json:"lir_satellite_image_name"`
	FiscoBcosImageName          string `json:"fisco_bcos_image_name"`
}
