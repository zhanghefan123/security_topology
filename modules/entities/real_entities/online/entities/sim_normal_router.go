package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/decider"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimNormalRouter struct {
	*SimNodeBase
	StartCorruptRatio              float64
	EndCorruptRatio                float64
	StartCorruptSpecialPacketRatio float64
	EndCorruptSpecialPacketRatio   float64
	CorruptDecider                 *decider.ActionDecider // 可以是任何形式的丢包分布
	CorruptSpecialDecider          *decider.ActionDecider // 可以是任何形式的丢包分布
}

// NewSimNormalRouter 创建新的路由器
func NewSimNormalRouter(nodeName string, nodeIndex int,
	startCorruptRatio, endCorruptRatio, startCorruptSpecialPacketRatio, endCorruptSpecialPacketRatio float64) (*SimNormalRouter, error) {
	simNodeBase := CreateSimNodeBase(nodeName, nodeIndex)
	var corruptDecider *decider.ActionDecider
	var corruptSpecialDecider *decider.ActionDecider
	var err error
	corruptDecider, err = decider.CreateUniformDecider(startCorruptRatio, endCorruptRatio, nodeIndex)
	if err != nil {
		return nil, fmt.Errorf("create uniform corrupt decider failed: %v", err)
	}
	corruptSpecialDecider, err = decider.CreateUniformDecider(startCorruptSpecialPacketRatio, endCorruptSpecialPacketRatio, nodeIndex)
	return &SimNormalRouter{
		SimNodeBase:           simNodeBase,
		StartCorruptRatio:     startCorruptRatio,
		EndCorruptRatio:       endCorruptRatio,
		CorruptDecider:        corruptDecider,
		CorruptSpecialDecider: corruptSpecialDecider,
	}, nil
	// the purpose of the attacker is to let the source select the path and they manipulate the delay
}

func (router *SimNormalRouter) Reset(startCorruptRatio, endCorruptRatio, startCorruptSpecialRatio, endCorruptSpecialRatio float64) error {
	var corruptDecider *decider.ActionDecider
	var corruptSpecialDecider *decider.ActionDecider
	var err error
	corruptDecider, err = decider.CreateUniformDecider(startCorruptRatio, endCorruptRatio, router.Index)
	if err != nil {
		return fmt.Errorf("CreateUniformDecider failed: %v", err)
	}
	corruptDecider, err = decider.CreateUniformDecider(startCorruptSpecialRatio, endCorruptSpecialRatio, router.Index)
	if err != nil {
		return fmt.Errorf("create uniform corrupt decider failed: %v", err)
	}
	router.StartCorruptRatio = startCorruptRatio
	router.EndCorruptRatio = endCorruptRatio
	router.CorruptDecider = corruptDecider
	router.CorruptSpecialDecider = corruptSpecialDecider
	return nil
}

// TakeActionOnPacket 按照策略进行数据包的篡改
func (router *SimNormalRouter) TakeActionOnPacket(pkt *SimPacket) error {
	if pkt.Type == types.SimPacketType_DataPacket {
		// 根据概率决策是否将包进行篡改
		if router.CorruptDecider.ShouldTakeAction() {
			pkt.IsCorrupted = true
		}
	} else if pkt.Type == types.SimPacketType_AckPacket {
		// 根据概率决定是否将包进行篡改
		if router.CorruptSpecialDecider.ShouldTakeAction() {
			pkt.IsCorrupted = true
		}
	} else {
		return fmt.Errorf("unsupported packet type")
	}
	return nil
}

// ProcessPacket 进行数据包的处理
func (router *SimNormalRouter) ProcessPacket(pkt *SimPacket) error {
	// 1. data packet 的处理逻辑
	if pkt.Type == types.SimPacketType_DataPacket {
		if pkt.IsCorrupted {
			return fmt.Errorf("the packet is already dropped or corrupted")
		} else {
			err := router.TakeActionOnPacket(pkt)
			if err != nil {
				return fmt.Errorf("take action on data packet failed due to: %w", err)
			}
		}
	} else if pkt.Type == types.SimPacketType_AckPacket {
		// 2. ack packet 的处理逻辑
		err := router.TakeActionOnPacket(pkt)
		if err != nil {
			return fmt.Errorf("take action on data packet failed due to: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported packet type")
	}

	return nil
}
