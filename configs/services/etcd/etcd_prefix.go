package etcd

type EtcdPrefix struct {
	SatellitesPrefix     string `mapstructure:"satellites_prefix"`
	GroundStationsPrefix string `mapstructure:"ground_stations_prefix"`
	ISLsPrefix           string `mapstructure:"isls_prefix"`
	GSLsPrefix           string `mapstructure:"gsls_prefix"`
}
