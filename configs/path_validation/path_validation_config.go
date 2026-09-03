package path_validation

type MultipathConfig struct {
	MultipathRoutingType int    `mapstructure:"multipath_routing_type"`
	ValidationTopology   string `mapstructure:"validation_topology"`
	MultipathFileName    string `mapstructure:"multipath_file_name"`
}

type SecPathMabConfig struct {
	SecPathMabType         int // 实验的类型
	NumberOfHops           int // 端到端的跳数
	NumberOfSegmentsPerHop int // 每跳的 segment 数量
	ExperimentType         int // 实验的类型
	TopologyType           int // 拓扑的类型
}

type PathValidationConfig struct {
	BfEffectiveBits            int              `mapstructure:"bf_effective_bits"`
	PVFEffectiveBits           int              `mapstructure:"pvf_effective_bits"`
	HashSeed                   int              `mapstructure:"hash_seed"`
	NumberOfHashFunctions      int              `mapstructure:"number_of_hash_functions"`
	LiRSingleTimeEncodingCount int              `mapstructure:"lir_single_time_encoding_count"`
	RoutingTableType           int              `mapstructure:"routing_table_type"`
	TransmissionType           int              `mapstructure:"transmission_type"`
	MultipathConfig            MultipathConfig  `mapstructure:"multipath_config"`
	SecPathMabConfig           SecPathMabConfig `mapstructure:"sec_path_mab_config"`
}
