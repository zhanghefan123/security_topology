package network

type NetworkConfig struct {
	LocalNetworkAddress string `mapstructure:"local_network_address"`
	BaseNetworkAddress  string `mapstructure:"base_network_address"`
}
