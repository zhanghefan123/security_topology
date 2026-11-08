package fisco_bcos

type FiscoBcosParams struct {
	ProjectPath                       string `json:"project_path"`
	ExamplePath                       string `json:"example_path"`
	ConsolePath                       string `json:"console_path"`
	P2pStartPort                      int    `json:"p2p_start_port"`
	RpcStartPort                      int    `json:"rpc_start_port"`
	LeaderPeriod                      int    `json:"leader_period"`
	ConsensusTimeout                  int    `json:"consensus_timeout"`
	SealWatiMs                        int    `json:"sealer_wait_ms"`
	MinSealTime                       int    `json:"min_seal_time"`
	AllowedLaggingBehindBlocks        int    `json:"allowed_lagging_behind_blocks"`
	WaterMarkLimit                    int    `json:"water_mark_limit"`
	LogLevel                          string `json:"log_level"`
	BlockTxCountLimit                 int    `json:"block_tx_count_limit"`
	RecursiveTrigger                  bool   `json:"recursive_trigger"`
	FastTrigger                       bool   `json:"fast_trigger"`
	LargeInterval                     bool   `json:"large_interval"`
	UseModifiedRequestBlocks          bool   `json:"use_modified_request_blocks"`
	EnablePending                     bool   `json:"enable_pending"`
	SyncIdleWait                      int    `json:"sync_idle_wait"`
	DownloadBlockProcessorThreadCount int    `json:"download_block_processor_thread_count"`
	SendBlockProcessorThreadCount     int    `json:"send_block_processor_thread_count"`
	MaxShardPerPeer                   int    `json:"max_shard_per_peer"`
	MaxBlocksPerRequest               int    `json:"max_blocks_per_request"`
	RequestPeerLimit                  int    `json:"request_peer_limit"`
	ExpectedTTL                       int    `json:"expected_ttl"`
	SyncSleepMs                       int    `json:"sync_sleep_ms"`
	EnableBlackList                   bool   `json:"enable_blacklist"`
	BlockIntervalMs                   int    `json:"block_interval_ms"`
	NewViewWaitMs                     int    `json:"new_view_wait_ms"`
}
