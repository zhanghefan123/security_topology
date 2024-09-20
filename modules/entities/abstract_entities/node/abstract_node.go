package node

import (
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/modules/logger"
)

var moduleAbstractEntities = logger.GetLogger(logger.ModuleAbstractEntities)

type AbstractNode struct {
	Type       types.NetworkNodeType // 节点类型
	ActualNode interface{}           // 实际的节点
}
