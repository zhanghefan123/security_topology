package chainmaker

type ChainMakerConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	P2pStartPort    int    `mapstructure:"p2p_start_port"`
	RpcStartPort    int    `mapstructure:"rpc_start_port"`
	HttpStartPort   int    `mapstructure:"http_start_port"`
	ConsensusType   int    `mapstructure:"consensus_type"`
	LogLevel        string `mapstructure:"log_level"`
	VmGoRuntimePort int    `mapstructure:"vm_go_runtime_port"`
	VmGoEnginePort  int    `mapstructure:"vm_go_engine_port"`

	EnableBroadcastDefence   bool    `mapstructure:"enable_broadcast_defence"`
	DirectRemoveAttackedNode bool    `mapstructure:"direct_remove_attacked_node"`
	SpeedCheck               bool    `mapstructure:"speed_check"`
	BlocksPerProposer        int     `mapstructure:"blocks_per_proposer"`
	DdosWarningRate          float64 `mapstructure:"ddos_warning_rate"`

	// 一些路径相关的配置
	ChainMakerGoProjectPath string `mapstructure:"chainmaker_go_project_path"`
	ChainMakerBuild         string `mapstructure:"chainmaker_build"`
	CryptoGenProjectPath    string `mapstructure:"crypto_gen_path"`
	TemplatesFilePath       string `mapstructure:"templates_file_path"`
}
