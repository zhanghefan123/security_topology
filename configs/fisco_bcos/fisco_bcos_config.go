package fisco_bcos

type FiscoBcosConfig struct {
	ProjectPath                       string `mapstructure:"project_path"`
	ExamplePath                       string `mapstructure:"example_path"`
	ConsolePath                       string `mapstructure:"console_path"`
	P2pStartPort                      int    `mapstructure:"p2p_start_port"`
	RpcStartPort                      int    `mapstructure:"rpc_start_port"`
	LeaderPeriod                      int    `mapstructure:"leader_period"`
	ConsensusTimeout                  int    `mapstructure:"consensus_timeout"`
	SealWatiMs                        int    `mapstructure:"sealer_wait_ms"`
	MinSealTime                       int    `mapstructure:"min_seal_time"`
	AllowedLaggingBehindBlocks        int    `mapstructure:"allowed_lagging_behind_blocks"`
	WaterMarkLimit                    int    `mapstructure:"water_mark_limit"`
	LogLevel                          string `mapstructure:"log_level"`
	BlockTxCountLimit                 int    `mapstructure:"block_tx_count_limit"`
	RecursiveTrigger                  bool   `mapstructure:"recursive_trigger"`
	FastTrigger                       bool   `mapstructure:"fast_trigger"`
	LargeInterval                     bool   `mapstructure:"large_interval"`
	UseModifiedRequestBlocks          bool   `mapstructure:"use_modified_request_blocks"`
	EnablePending                     bool   `mapstructure:"enable_pending"`
	SyncIdleWait                      int    `mapstructure:"sync_idle_wait"`
	DownloadBlockProcessorThreadCount int    `mapstructure:"download_block_processor_thread_count"`
	SendBlockProcessorThreadCount     int    `mapstructure:"send_block_processor_thread_count"`
	MaxShardPerPeer                   int    `mapstructure:"max_shard_per_peer"`
	MaxBlocksPerRequest               int    `mapstructure:"max_blocks_per_request"`
	RequestPeerLimit                  int    `mapstructure:"request_peer_limit"`
	ExpectedTTL                       int    `mapstructure:"expected_ttl"`
	SyncSleepMs                       int    `mapstructure:"sync_sleep_ms"`
	EnableBlackList                   bool   `mapstructure:"enable_blacklist"`
	BlockIntervalMs                   int    `mapstructure:"block_interval_ms"`
	NewViewWaitMs                     int    `mapstructure:"new_view_wait_ms"`
}
