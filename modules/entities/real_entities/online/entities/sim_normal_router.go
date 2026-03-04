package entities

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/online/corrupt_decider"
)

type SimNormalRouter struct {
	*SimNodeBase
	CorruptDecider *corrupt_decider.CorruptDecider // 可以是任何形式的丢包分布
}

// NewSimNormalRouter 创建新的路由器
func NewSimNormalRouter(NodeName string, NodeIndex int, corruptDecider *corrupt_decider.CorruptDecider) *SimNormalRouter {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimNormalRouter{
		SimNodeBase:    simNodeBase,
		CorruptDecider: corruptDecider,
	}
}

// CorruptPacket 按照策略进行数据包的篡改
func (router *SimNormalRouter) CorruptPacket(pkt *SimPacket) {
	// 根据概率决策是否将包进行篡改
	if pkt.IsCorrupted {
		return
	} else {
		if router.CorruptDecider.ShouldCorrupt() {
			pkt.IsCorrupted = true
		}
	}
}

// ProcessPacket 进行数据包的处理
func (router *SimNormalRouter) ProcessPacket(pkt *SimPacket) {
	if pkt.IsCorrupted {
		return
	} else {
		router.CorruptPacket(pkt)
	}
}
