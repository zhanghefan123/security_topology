package fabric

type FabricConfig struct {
	OrderGeneralListenStartPort   int    `mapstructure:"order_general_listen_start_port"`
	OrderAdminListenStartPort     int    `mapstructure:"order_admin_listen_start_port"`
	OrderOperationListenStartPort int    `mapstructure:"order_operation_listen_start_port"`
	OrderStartPort                int    `mapstructure:"order_start_port"`
	PeerListenStartPort           int    `mapstructure:"peer_listen_start_port"`
	PeerChaincodeStartPort        int    `mapstructure:"peer_chaincode_start_port"`
	PeerOperationStartPort        int    `mapstructure:"peer_operation_start_port"`
	PprofOrdererStartListenPort   int    `mapstructure:"pprof_orderer_start_listen_port"`
	PprofPeerStartListenPort      int    `mapstructure:"pprof_peer_start_listen_port"`
	EnablePprof                   bool   `mapstructure:"enable_pprof"`
	EnableRoutine                 bool   `mapstructure:"enable_routine"`
	EnableAdvancedMessageHandler  bool   `mapstructure:"enable_advanced_message_handler"`
	ConsensusType                 int    `mapstructure:"consensus_type"`
	LogLevel                      string `mapstructure:"log_level"`
	EnableBroadcastDefence        bool   `mapstructure:"enable_broadcast_defence"`
	DirectRemoveAttackedNode      bool   `mapstructure:"direct_remove_attacked_node"`
	SpeedCheck                    bool   `mapstructure:"speed_check"`
	FabricProjectPath             string `mapstructure:"fabric_project_path"`
	FabricSamplesPath             string `mapstructure:"fabric_samples_path"`
	FabricNetworkPath             string `mapstructure:"fabric_network_path"`
}
