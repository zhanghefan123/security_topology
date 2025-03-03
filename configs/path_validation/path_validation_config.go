package path_validation

type PathValidationConfig struct {
	RoutingTableType           int `mapstructure:"routing_table_type"`
	BfEffectiveBits            int `mapstructure:"bf_effective_bits"`
	PVFEffectiveBits           int `mapstructure:"pvf_effective_bits"`
	HashSeed                   int `mapstructure:"hash_seed"`
	NumberOfHashFunctions      int `mapstructure:"number_of_hash_functions"`
	LiRSingleTimeEncodingCount int `mapstructure:"lir_single_time_encoding_count"`
}
