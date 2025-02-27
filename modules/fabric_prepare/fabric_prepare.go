package fabric_prepare

import "zhanghefan123/security_topology/modules/logger"

var (
	fabricPrepareWorkLogger = logger.GetLogger(logger.ModuleFabricPrepare)
)

type FabricPrepare struct {
	fabricPeerNodeCount  int
	fabricOrderNodeCount int
	generateSteps        map[string]struct{}
	pathMapping          map[string]string
}

func NewFabricPrepare(fabricPeerNodeCount, fabricOrderNodeCount int) *FabricPrepare {
	return &FabricPrepare{
		fabricPeerNodeCount:  fabricPeerNodeCount,
		fabricOrderNodeCount: fabricOrderNodeCount,
		generateSteps:        make(map[string]struct{}),
		pathMapping:          make(map[string]string),
	}
}
