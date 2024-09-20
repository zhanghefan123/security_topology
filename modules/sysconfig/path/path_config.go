package path

type PathConfig struct {
	ResourcesPath      string  `mapstructure:"resources_path"`
	ConfigGeneratePath string  `mapstructure:"config_generate_path"`
	FrrPath            FrrPath `mapstructure:"frr_path"`
}
