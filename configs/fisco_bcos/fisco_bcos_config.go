package fisco_bcos

type FiscoBcosConfig struct {
	ProjectPath  string `mapstructure:"project_path"`
	ExamplePath  string `mapstructure:"example_path"`
	ConsolePath  string `mapstructure:"console_path"`
	P2pStartPort int    `mapstructure:"p2p_start_port"`
	RpcStartPort int    `mapstructure:"rpc_start_port"`
	LeaderPeriod int    `mapstructure:"leader_period"`
}
