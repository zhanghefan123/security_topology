package services

type EtcdPrefix struct {
	SatellitesPrefix     string `json:"satellites_prefix"`
	GroundStationsPrefix string `json:"ground_stations_prefix"`
	ISLsPrefix           string `json:"isls_prefix"`
	GSLsPrefix           string `json:"gsls_prefix"`
}

type EtcdConfig struct {
	ClientPort int        `json:"client_port"`
	PeerPort   int        `json:"peer_port"`
	DataDir    string     `json:"data_dir"`
	EtcdName   string     `json:"etcd_name"`
	EtcdPrefix EtcdPrefix `json:"etcd_prefix"`
}

type WebConfig struct {
	StartPort int `json:"start_port"`
}

type PositionUpdateConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
}

type DelayUpdateConfig struct {
	Enabled bool `json:"enabled"`
}

type Params struct {
	EtcdConfig           EtcdConfig           `json:"etcd_config"`
	WebConfig            WebConfig            `json:"web_config"`
	PositionUpdateConfig PositionUpdateConfig `json:"position_update_config"`
	DelayUpdateConfig    DelayUpdateConfig    `json:"delay_update_config"`
}
