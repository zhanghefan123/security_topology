package network

type NetworkConfig struct {
	LocalNetworkAddress  string
	BaseV4NetworkAddress string `mapstructure:"base_v4_network_address"`
	BaseV6NetworkAddress string `mapstructure:"base_v6_network_address"`
	HttpListenPort       string `mapstructure:"http_listen_port"`
	EnableFrr            bool   `mapstructure:"enable_frr"`
	OspfVersion          string `mapstructure:"ospf_version"`
	EnableSRv6           bool   `mapstructure:"enable_srv6"`
}
