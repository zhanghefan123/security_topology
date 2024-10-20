package topology

type TopologyConfig struct {
	RouterImageName        string `mapstructure:"router_image_name"`
	NormalNodeImageName    string `mapstructure:"normal_node_image_name"`
	ConsensusNodeImageName string `mapstructure:"consensus_node_image_name"`
	MaliciousNodeImageName string `mapstructure:"malicious_node_image_name"`
}
