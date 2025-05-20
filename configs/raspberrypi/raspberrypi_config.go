package raspberrypi

type RaspberryPiConfig struct {
	NodeIDs     []int    `mapstructure:"node_ids"`
	NodeTypes   []string `mapstructure:"node_types"`
	IpAddresses []string `mapstructure:"ip_addresses"`
	Connections []string `mapstructure:"connections"`
	UserName    string   `mapstructure:"user_name"`
	Password    string   `mapstructure:"password"`
	GrpcPort    int      `mapstructure:"grpc_port"`
	PythonPath  string   `mapstructure:"python_path"`
}
