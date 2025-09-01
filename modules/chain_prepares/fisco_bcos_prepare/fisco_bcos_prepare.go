package fisco_bcos_prepare

import "zhanghefan123/security_topology/modules/logger"

var (
	fiscoBcosPrepareWorkLogger = logger.GetLogger(logger.ModuleFiscoBcosPrepare)
)

type FiscoBcosPrepare struct {
	fiscoBcosNodeCount int                 // fisco bcos 节点数量
	firstIpAddresses   []string            // 每个容器的首地址
	p2pStartPort       int                 // p2p 起始端口
	generateSteps      map[string]struct{} // 执行的步骤
}

// NewFiscoBcosPrepare 新的 BcosPrepare
func NewFiscoBcosPrepare(fiscoBcosNodeCount int, firstIpAddresses []string, p2pStartPort int) *FiscoBcosPrepare {
	return &FiscoBcosPrepare{
		fiscoBcosNodeCount: fiscoBcosNodeCount,
		firstIpAddresses:   firstIpAddresses,
		p2pStartPort:       p2pStartPort,
		generateSteps:      make(map[string]struct{}),
	}
}
