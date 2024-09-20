package path

type FrrPath struct {
	FrrHostPath      string `mapstructure:"frr_host_path"`
	FrrContainerPath string `mapstructure:"frr_container_path"`
}
