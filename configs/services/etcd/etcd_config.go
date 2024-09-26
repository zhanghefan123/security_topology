package etcd

type EtcdConfig struct {
	ClientPort int        `mapstructure:"client_port"`
	PeerPort   int        `mapstructure:"peer_port"`
	DataDir    string     `mapstructure:"data_dir"`
	EtcdName   string     `mapstructure:"etcd_name"`
	ImageName  string     `mapstructure:"image_name"`
	EtcdPrefix EtcdPrefix `mapstructure:"etcd_prefix"`
}
