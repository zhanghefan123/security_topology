package lir

type LiRConfig struct {
	EffectiveBits         int `mapstructure:"effective_bits"`
	HashSeed              int `mapstructure:"hash_seed"`
	NumberOfHashFunctions int `mapstructure:"number_of_hash_functions"`
}
