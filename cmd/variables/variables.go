package variables

const (
	ImageUbuntuWithSoftware = "ubuntu_with_software"
	ImagePythonEnv          = "python_env"
	ImageGoEnv              = "go_env"
	ImageNormalSatellite    = "normal_satellite"
	ImageNameEtcd           = "etcd_service"
	ImageNamePosition       = "position_service"
	AllImages               = "all_images"
)

const (
	OperationBuild   string = "build"
	OperationRebuild string = "rebuild"
	OperationRemove  string = "remove"
)

var (
	UserSelectedImage     = ImageNormalSatellite
	UserSelectedOperation = OperationBuild
	ImagesInBuildOrder    = []string{ImageUbuntuWithSoftware, ImageGoEnv, ImageNormalSatellite}
	AvailableOperations   = map[string]interface{}{
		OperationBuild:   struct{}{},
		OperationRebuild: struct{}{},
		OperationRemove:  struct{}{},
	}
	ExistedImages = map[string]bool{
		ImageUbuntuWithSoftware: false,
		ImagePythonEnv:          false,
		ImageGoEnv:              false,
		ImageNormalSatellite:    false,
		ImageNameEtcd:           false,
		ImageNamePosition:       false,
		AllImages:               false,
	}
)
