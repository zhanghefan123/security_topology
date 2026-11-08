package resources

type Params struct {
	Enabled       bool    `json:"enabled"`
	CpuLimit      float64 `json:"cpu_limit"`
	MemoryLimit   float64 `json:"memory_limit"`
	TopologiesDir string  `json:"topologies_dir"`
}
