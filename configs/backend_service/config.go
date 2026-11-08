package backend_service

type BackendServiceConfig struct {
	HttpListenPort      string `mapstructure:"http_listen_port"`
	LocalNetworkAddress string `mapstructure:"local_network_address"`
}
