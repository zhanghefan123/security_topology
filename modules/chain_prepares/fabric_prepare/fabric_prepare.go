package fabric_prepare

import "zhanghefan123/security_topology/modules/logger"

var (
	fabricPrepareWorkLogger = logger.GetLogger(logger.ModuleFabricPrepare)
)

type FabricPrepare struct {
	fabricPeerNodeCount    int // 对等节点数量
	fabricOrdererNodeCount int // 排序节点数量
	generateSteps          map[string]struct{}
	pathMapping            map[string]string
}

// NewFabricPrepare 创建 FabricPrepare 对象
func NewFabricPrepare(fabricPeerNodeCount, fabricOrdererNodeCount int) *FabricPrepare {
	return &FabricPrepare{
		fabricPeerNodeCount:    fabricPeerNodeCount,
		fabricOrdererNodeCount: fabricOrdererNodeCount,
		generateSteps:          make(map[string]struct{}),
		pathMapping:            make(map[string]string),
	}
}
