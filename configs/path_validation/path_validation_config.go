package path_validation

type PathValidationConfig struct {
	RoutingTableType      int `mapstructure:"routing_table_type"`
	EffectiveBits         int `mapstructure:"effective_bits"`
	HashSeed              int `mapstructure:"hash_seed"`
	NumberOfHashFunctions int `mapstructure:"number_of_hash_functions"`
}
