package etcd

type EtcdPrefix struct {
	SatellitesPrefix string `mapstructure:"satellites_prefix"`
	ISLsPrefix       string `mapstructure:"isls_prefix"`
}
