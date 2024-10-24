package chainmaker_prepare

import (
	"zhanghefan123/security_topology/modules/logger"
)

var (
	prepareWorkLogger = logger.GetLogger(logger.ModulePrepare)
)

type ChainMakerPrepare struct {
	nodeCount     int
	peerIdList    []string
	generateSteps map[string]struct{}
	pathMapping   map[string]string
	ipv4Addresses []string
}

// NewChainMakerPrepare 创建新的 chainmakerPrepare
func NewChainMakerPrepare(nodeCount int, ipv4Addresses []string) *ChainMakerPrepare {
	return &ChainMakerPrepare{
		nodeCount:     nodeCount,
		peerIdList:    make([]string, 0),
		generateSteps: make(map[string]struct{}),
		pathMapping:   make(map[string]string),
		ipv4Addresses: ipv4Addresses,
	}
}
