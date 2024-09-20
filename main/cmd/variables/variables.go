package variables

const (
	ImageUbuntuWithSoftware = "ubuntu_with_software"
	ImagePythonEnv          = "python_env"
	ImageGoEnv              = "go_env"
	ImageNormalSatellite    = "normal_satellite"
)

const (
	OperationBuild   string = "build"
	OperationRebuild string = "rebuild"
	OperationRemove  string = "remove"
)

var (
	UserSelectedImage     string                 = ImageNormalSatellite
	UserSelectedOperation string                 = OperationBuild
	AvailableOperations   map[string]interface{} = map[string]interface{}{
		OperationBuild:   struct{}{},
		OperationRebuild: struct{}{},
		OperationRemove:  struct{}{},
	}
	ExistedImages map[string]bool = map[string]bool{
		ImageUbuntuWithSoftware: false,
		ImagePythonEnv:          false,
		ImageGoEnv:              false,
		ImageNormalSatellite:    false,
	}
)
