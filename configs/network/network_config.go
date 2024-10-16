package network

type NetworkConfig struct {
	LocalNetworkAddress  string
	BaseV4NetworkAddress string `mapstructure:"base_v4_network_address"`
	BaseV6NetworkAddress string `mapstructure:"base_v6_network_address"`
	EnableFrr            bool   `mapstructure:"enable_frr"`
}
