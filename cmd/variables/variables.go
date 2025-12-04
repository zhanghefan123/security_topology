package variables

const (
	ImageNameUbuntuWithSoftware = "ubuntu_with_software"
	ImageNamePythonEnv          = "python_env"
	ImageNameGoEnv              = "go_env"
	ImageNameNormalSatellite    = "normal_satellite"
	ImageNameEtcd               = "etcd_service"
	ImageNameRealTimePosition   = "realtime_position"
	ImageNameRouter             = "router"
	ImageNameNormalNode         = "normal_node"
	ImageNameGroundStation      = "ground_station"
	ImageNameConsensusNode      = "consensus_node"
	ImageNameMaliciousNode      = "malicious_node"
	ImageNameLiRNode            = "lir_node"
	ImageNameChainMakerEnv      = "chainmaker_env"
	ImageNameChainMaker         = "chainmaker"
	ImageNameFabricOrder        = "fabric-orderer"
	ImageNameFabricPeer         = "fabric-peer"
	ImageNameFiscoBcos          = "fiscoorg/fiscobcos"
	AllImages                   = "all_images"
)

// 基于官方镜像的镜像 -> etcd_service
// 基于 ubuntu_with_software  的镜像 -> python_env, go_env, chainmaker_env,
// fabric 的镜像 -> fabric-orderer fabric-peer
// 基于 go_env 的镜像 -> normal_satellite, router, consensus_node, normal_node, ground_station
// 基于 python_env 的镜像 -> lir_node, malicious_node, realtime_position
// 基于 chainmaker_env 的镜像 -> chainmaker

const (
	OperationBuild   string = "build"
	OperationRebuild string = "rebuild"
	OperationRemove  string = "remove"
)

var (
	UserSelectedImage            = ImageNameEtcd
	UserSelectedOperation        = OperationBuild
	UserSelectedExperimentNumber = 1
	ImagesInBuildOrder           = []string{
		ImageNameEtcd,
		ImageNameUbuntuWithSoftware,
		ImageNameFabricOrder, ImageNameFabricPeer,
		ImageNamePythonEnv, ImageNameGoEnv, ImageNameChainMakerEnv,
		ImageNameNormalNode, ImageNameGroundStation, ImageNameRouter, ImageNameConsensusNode, ImageNameNormalSatellite, ImageNameChainMaker,
		ImageNameMaliciousNode, ImageNameLiRNode, ImageNameRealTimePosition, ImageNameFiscoBcos,
	}
	AvailableOperations = map[string]interface{}{
		OperationBuild:   struct{}{},
		OperationRebuild: struct{}{},
		OperationRemove:  struct{}{},
	}
	AvailableImages = map[string]interface{}{
		ImageNameEtcd: struct{}{},

		ImageNameUbuntuWithSoftware: struct{}{},
		ImageNamePythonEnv:          struct{}{},
		ImageNameGoEnv:              struct{}{},
		ImageNameChainMakerEnv:      struct{}{},

		ImageNameFabricOrder: struct{}{},
		ImageNameFabricPeer:  struct{}{},

		ImageNameGroundStation:   struct{}{},
		ImageNameNormalNode:      struct{}{},
		ImageNameRouter:          struct{}{},
		ImageNameConsensusNode:   struct{}{},
		ImageNameNormalSatellite: struct{}{},

		ImageNameMaliciousNode:    struct{}{},
		ImageNameLiRNode:          struct{}{},
		ImageNameRealTimePosition: struct{}{},

		ImageNameChainMaker: struct{}{},

		ImageNameFiscoBcos: struct{}{},

		AllImages: struct{}{},
	}

	ExistedImages = map[string]bool{
		ImageNameEtcd: false,

		ImageNameUbuntuWithSoftware: false,
		ImageNamePythonEnv:          false,
		ImageNameGoEnv:              false,
		ImageNameChainMakerEnv:      false,

		ImageNameFabricOrder: false,
		ImageNameFabricPeer:  false,

		ImageNameGroundStation:   false,
		ImageNameNormalNode:      false,
		ImageNameRouter:          false,
		ImageNameConsensusNode:   false,
		ImageNameNormalSatellite: false,

		ImageNameMaliciousNode:    false,
		ImageNameLiRNode:          false,
		ImageNameRealTimePosition: false,

		ImageNameChainMaker: false,
		ImageNameFiscoBcos:  false,
	}
)
