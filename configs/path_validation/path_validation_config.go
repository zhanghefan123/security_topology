package path_validation

type MultipathConfig struct {
	MultipathRoutingType int    `mapstructure:"multipath_routing_type"`
	ValidationTopology   string `mapstructure:"validation_topology"`
	MultipathFileName    string `mapstructure:"multipath_file_name"`
}

type PathValidationConfig struct {
	BfEffectiveBits            int             `mapstructure:"bf_effective_bits"`
	PVFEffectiveBits           int             `mapstructure:"pvf_effective_bits"`
	HashSeed                   int             `mapstructure:"hash_seed"`
	NumberOfHashFunctions      int             `mapstructure:"number_of_hash_functions"`
	LiRSingleTimeEncodingCount int             `mapstructure:"lir_single_time_encoding_count"`
	RoutingTableType           int             `mapstructure:"routing_table_type"`
	TransmissionType           int             `mapstructure:"transmission_type"`
	MultipathConfig            MultipathConfig `mapstructure:"multipath_config"`
	SecPathMabType             int
	PerLinkDelay               float64
}
