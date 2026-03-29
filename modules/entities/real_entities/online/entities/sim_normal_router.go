package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/decider"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimNormalRouter struct {
	*SimNodeBase
	StartDropRatio                 float64
	EndDropRatio                   float64
	StartCorruptRatio              float64
	EndCorruptRatio                float64
	StartCorruptSpecialPacketRatio float64
	EndCorruptSpecialPacketRatio   float64
	DropDecider                    *decider.ActionDecider // 可以是任何形式的丢包分布
	CorruptDecider                 *decider.ActionDecider // 可以是任何形式的丢包分布
	CorruptSpecialPacketDecider    *decider.ActionDecider // 可以是任何形式的丢包分布
}

// NewSimNormalRouter 创建新的路由器
func NewSimNormalRouter(nodeName string, nodeIndex int, startDropRatio, endDropRatio,
	startCorruptRatio, endCorruptRatio, startCorruptSpecialPacketRatio, endCorruptSpecialPacketRatio float64) (*SimNormalRouter, error) {
	simNodeBase := CreateSimNodeBase(nodeName, nodeIndex)
	var corruptDecider *decider.ActionDecider
	var dropDecider *decider.ActionDecider
	var corruptSpecialPacketDecider *decider.ActionDecider
	var err error
	dropDecider, err = decider.CreateUniformDecider(startDropRatio, endDropRatio, nodeIndex)
	if err != nil {
		return nil, fmt.Errorf("CreateUniformDecider failed: %v", err)
	}
	corruptDecider, err = decider.CreateUniformDecider(startCorruptRatio, endCorruptRatio, nodeIndex)
	if err != nil {
		return nil, fmt.Errorf("create uniform corrupt decider failed: %v", err)
	}
	corruptSpecialPacketDecider, err = decider.CreateUniformDecider(startCorruptSpecialPacketRatio, endCorruptSpecialPacketRatio, nodeIndex)
	return &SimNormalRouter{
		SimNodeBase:                 simNodeBase,
		StartDropRatio:              startDropRatio,
		EndDropRatio:                endDropRatio,
		StartCorruptRatio:           startCorruptRatio,
		EndCorruptRatio:             endCorruptRatio,
		DropDecider:                 dropDecider,
		CorruptDecider:              corruptDecider,
		CorruptSpecialPacketDecider: corruptSpecialPacketDecider,
	}, nil
	// the purpose of the attacker is to let the source select the path and they manipulate the delay
}

func (router *SimNormalRouter) Reset(startDropRatio, endDropRatio, startCorruptRatio, endCorruptRatio float64) error {
	var corruptDecider *decider.ActionDecider
	var dropDecider *decider.ActionDecider
	var err error
	dropDecider, err = decider.CreateUniformDecider(startDropRatio, endDropRatio, router.Index)
	if err != nil {
		return fmt.Errorf("CreateUniformDecider failed: %v", err)
	}
	corruptDecider, err = decider.CreateUniformDecider(startCorruptRatio, endCorruptRatio, router.Index)
	if err != nil {
		return fmt.Errorf("create uniform corrupt decider failed: %v", err)
	}
	router.StartDropRatio = startDropRatio
	router.EndDropRatio = endDropRatio
	router.StartCorruptRatio = startCorruptRatio
	router.EndCorruptRatio = endCorruptRatio
	router.DropDecider = dropDecider
	router.CorruptDecider = corruptDecider
	return nil
}

// TakeActionOnPacket 按照策略进行数据包的篡改
func (router *SimNormalRouter) TakeActionOnPacket(pkt *SimPacket) {
	// 根据概率决策是否将包进行篡改
	if router.DropDecider.ShouldTakeAction() {
		//fmt.Println("drop packet")
		pkt.IsDropped = true
		return
	} else {
		//fmt.Println("not drop packet")
		if router.CorruptDecider.ShouldTakeAction() {
			pkt.IsCorrupted = true
		}
	}
}

// ProcessPacket 进行数据包的处理
func (router *SimNormalRouter) ProcessPacket(pkt *SimPacket) error {
	if pkt.Type == types.SimPacketType_DataPacket {
		if pkt.IsDropped || pkt.IsCorrupted {
			return fmt.Errorf("the packet is already dropped or corrupted")
		} else {
			// 1. 首先决定是否丢包
			router.TakeActionOnPacket(pkt)
		}
	} else if pkt.Type == types.SimPacketType_AckPacket {
		if pkt.IsDropped {
			return fmt.Errorf("the packet is already dropped or corrupted")
		} else {
			// 1. 首先决定是否丢包
			// router.TakeActionOnPacket(pkt)
			return nil
		}
	} else {
		return fmt.Errorf("unsupported packet type")
	}

	return nil
}
