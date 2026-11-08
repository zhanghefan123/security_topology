package fabric

type FabricParams struct {
	OrderGeneralListenStartPort   int    `json:"order_general_listen_start_port"`
	OrderAdminListenStartPort     int    `json:"order_admin_listen_start_port"`
	OrderOperationListenStartPort int    `json:"order_operation_listen_start_port"`
	OrderStartPort                int    `json:"order_start_port"`
	PeerListenStartPort           int    `json:"peer_listen_start_port"`
	PeerChaincodeStartPort        int    `json:"peer_chaincode_start_port"`
	PeerOperationStartPort        int    `json:"peer_operation_start_port"`
	PprofOrdererStartListenPort   int    `json:"pprof_orderer_start_listen_port"`
	PprofPeerStartListenPort      int    `json:"pprof_peer_start_listen_port"`
	EnablePprof                   bool   `json:"enable_pprof"`
	EnableRoutine                 bool   `json:"enable_routine"`
	EnableAdvancedMessageHandler  bool   `json:"enable_advanced_message_handler"`
	EnableLeaderStepDown          bool   `json:"enable_leader_step_down"`
	ConsensusType                 int    `json:"consensus_type"`
	LogLevel                      string `json:"log_level"`
	EnableBroadcastDefence        bool   `json:"enable_broadcast_defence"`
	DirectRemoveAttackedNode      bool   `json:"direct_remove_attacked_node"`
	SpeedCheck                    bool   `json:"speed_check"`
	FabricProjectPath             string `json:"fabric_project_path"`
	FabricSamplesPath             string `json:"fabric_samples_path"`
	FabricNetworkPath             string `json:"fabric_network_path"`
	ForwardTimeout                int    `json:"forward_timeout"`
	ComplainTimeout               int    `json:"complain_timeout"`
	BlockIntervalMs               int    `json:"block_interval_ms"`
	SealWaitMs                    int    `json:"seal_wait_ms"`
	WithDelay                     bool   `json:"with_delay"`
	RequestBatchMaxCount          int    `json:"request_batch_max_count"`
	RequestBatchMaxBytes          int    `json:"request_batch_max_bytes"`
}
