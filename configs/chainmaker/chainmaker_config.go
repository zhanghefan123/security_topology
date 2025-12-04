package chainmaker

type ChainMakerConfig struct {
	P2pStartPort    int    `mapstructure:"p2p_start_port"`
	RpcStartPort    int    `mapstructure:"rpc_start_port"`
	LogLevel        string `mapstructure:"log_level"`
	VmGoRuntimePort int    `mapstructure:"vm_go_runtime_port"`
	VmGoEnginePort  int    `mapstructure:"vm_go_engine_port"`

	EnableDefence   bool `mapstructure:"enable_defence"`
	CheckDdosPeriod int  `mapstructure:"check_ddos_period_ms"`

	BlocksPerProposer     int  `mapstructure:"blocks_per_proposer"`
	TimeoutPropose        int  `mapstructure:"timeout_propose"`
	TimeoutProposeOptimal int  `mapstructure:"timeout_propose_optimal"`
	ProposeOptimal        bool `mapstructure:"propose_optimal"`
	EnableBlackList       bool `mapstructure:"enable_blacklist"`

	// 一些路径相关的配置
	ChainMakerGoProjectPath string `mapstructure:"chainmaker_go_project_path"`
	ChainMakerBuild         string `mapstructure:"chainmaker_build"`
	TemplatesFilePath       string `mapstructure:"templates_file_path"`

	ResendSync     bool `mapstructure:"resend_sync"`
	TickIntervalMs int  `mapstructure:"tick_interval_ms"`

	BatchSizeFromOneNode int `mapstructure:"batch_size_from_one_node"`

	CryptoGenProjectPath string // $ChainMakerGoProjectPath/tools/chainmaker-cryptogen
	BlockReqTimeout      int    `mapstructure:"block_req_timeout"`
	BlockIntervalMs      int    `mapstructure:"block_interval_ms"`
	BlockTxCapacity      int    `mapstructure:"block_tx_capacity"`
}
