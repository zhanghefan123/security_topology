package variables

const (
	ImageUbuntuWithSoftware = "ubuntu_with_software"
	ImagePythonEnv          = "python_env"
	ImageGoEnv              = "go_env"
	ImageNormalSatellite    = "normal_satellite"
	ImageNameEtcd           = "etcd_service"
	ImageNamePosition       = "position_service"
	ImageNameRouter         = "router"
	ImageNameNormalNode     = "normal_node"
	ImageNameConsensusNode  = "consensus_node"
	ImageNameMaliciousNode  = "malicious_node"
	ImageLiRNode            = "lir_node"

	AllImages = "all_images"
)

const (
	OperationBuild   string = "build"
	OperationRebuild string = "rebuild"
	OperationRemove  string = "remove"
)

var (
	UserSelectedImage     = ImageNormalSatellite
	UserSelectedOperation = OperationBuild
	ImagesInBuildOrder    = []string{ImageUbuntuWithSoftware, ImagePythonEnv, ImageGoEnv,
		ImageNormalSatellite, ImageNameEtcd, ImageNamePosition, ImageNameRouter,
		ImageNameNormalNode, ImageNameConsensusNode, ImageNameMaliciousNode, ImageLiRNode}
	AvailableOperations = map[string]interface{}{
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
		ImageNameRouter:         false,
		ImageNameNormalNode:     false,
		ImageNameConsensusNode:  false,
		ImageNameMaliciousNode:  false,
		ImageLiRNode:            false,
		AllImages:               false,
	}
)
