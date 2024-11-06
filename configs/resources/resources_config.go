package resources

type ResourcesConfig struct {
	CpuLimit    float64 `mapstructure:"cpu_limit"`
	MemoryLimit float64 `mapstructure:"memory_limit"`
}
