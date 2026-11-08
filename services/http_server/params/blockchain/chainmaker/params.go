package chainmaker

type ChainMakerParams struct {
	P2pStartPort    int    `json:"p2p_start_port"`
	RpcStartPort    int    `json:"rpc_start_port"`
	LogLevel        string `json:"log_level"`
	VmGoRuntimePort int    `json:"vm_go_runtime_port"`
	VmGoEnginePort  int    `json:"vm_go_engine_port"`

	EnableDefence   bool `json:"enable_defence"`
	CheckDdosPeriod int  `json:"check_ddos_period_ms"`

	BlocksPerProposer     int  `json:"blocks_per_proposer"`
	TimeoutPropose        int  `json:"timeout_propose"`
	TimeoutProposeOptimal int  `json:"timeout_propose_optimal"`
	ProposeOptimal        bool `json:"propose_optimal"`
	EnableBlackList       bool `json:"enable_blacklist"`

	ChainMakerBuild   string `json:"chainmaker_build"`
	TemplatesFilePath string `json:"templates_file_path"`

	ResendSync     bool `json:"resend_sync"`
	TickIntervalMs int  `json:"tick_interval_ms"`

	BatchSizeFromOneNode int `json:"batch_size_from_one_node"`

	BlockReqTimeout int `json:"block_req_timeout"`
	BlockIntervalMs int `json:"block_interval_ms"`
	BlockTxCapacity int `json:"block_tx_capacity"`
}
