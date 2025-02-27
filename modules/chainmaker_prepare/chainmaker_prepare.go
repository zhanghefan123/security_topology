package chainmaker_prepare

import (
	"zhanghefan123/security_topology/modules/logger"
)

var (
	chainmakerPrepareWorkLogger = logger.GetLogger(logger.ModuleChainmakerPrepare)
)

var (
	chainmakerConsensusStrToIntMapping = map[string]int{
		"tbft": 1,
		"raft": 4,
	}
)

type ChainMakerPrepare struct {
	nodeCount             int
	peerIdList            []string
	generateSteps         map[string]struct{}
	pathMapping           map[string]string
	ipv4Addresses         []string
	selectedConsensusType int
}

// NewChainMakerPrepare 创建新的 chainmakerPrepare
func NewChainMakerPrepare(nodeCount int, ipv4Addresses []string, selectedConsensusType string) *ChainMakerPrepare {
	return &ChainMakerPrepare{
		nodeCount:             nodeCount,
		peerIdList:            make([]string, 0),
		generateSteps:         make(map[string]struct{}),
		pathMapping:           make(map[string]string),
		ipv4Addresses:         ipv4Addresses,
		selectedConsensusType: chainmakerConsensusStrToIntMapping[selectedConsensusType],
	}
}
