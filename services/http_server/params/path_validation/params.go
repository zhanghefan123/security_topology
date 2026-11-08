package path_validation

type MultipathParams struct {
	MultipathRoutingType int    `json:"multipath_routing_type"`
	ValidationTopology   string `json:"validation_topology"`
	MultipathFileName    string `json:"multipath_file_name"`
}

type SecPathMabParams struct {
	Enabled                bool    `json:"enabled"`
	SecPathMabType         int     `json:"sec_path_mab_type"`
	NumberOfHops           int     `json:"number_of_hops"`
	NumberOfSegmentsPerHop int     `json:"number_of_intermediate_nodes"`
	ExperimentType         int     `json:"experiment_type"`
	LowCorruptRatio        float64 `json:"low_corrupt_ratio"`
	HighCorruptRatio       float64 `json:"high_corrupt_ratio"`
}

type BasicParams struct {
	TransmissionType           int `json:"transmission_type"`
	BfEffectiveBits            int `json:"bf_effective_bits"`
	PVFEffectiveBits           int `json:"pvf_effective_bits"`
	HashSeed                   int `json:"hash_seed"`
	NumberOfHashFunctions      int `json:"number_of_hash_functions"`
	LiRSingleTimeEncodingCount int `json:"lir_single_time_encoding_count"`
	RoutingTableType           int `json:"routing_table_type"`
	SelirType                  int `json:"selir_type"`
	EpicType                   int `json:"epic_type"`
}

type PathValidationParams struct {
	BasicConfig      BasicParams      `json:"basic_config"`
	MultipathConfig  MultipathParams  `json:"multipath_config"`
	SecPathMabConfig SecPathMabParams `json:"sec_path_mab_config"`
}
