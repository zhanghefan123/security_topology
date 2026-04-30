package entities

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"

	"chainmaker.org/chainmaker/common/v2/random/uuid"
)

type SimPacket struct {
	Type        types.SimPacketType // 数据包的类型
	Uuid        string              // 数据包的唯一的标识
	SessionId   string              // 这个 packet 沿着哪个 session 传递
	IsCorrupted bool                // 是否已经被篡改了
	SampleNode  *SimAbstractNode    // 需要采样的 pv router
}

// CreateSimPacket 进行模拟数据包的创建
func CreateSimPacket(packetType types.SimPacketType, sessionId string, sampleNode *SimAbstractNode) *SimPacket {
	return &SimPacket{
		Type:        packetType,
		Uuid:        uuid.GetUUID(),
		SessionId:   sessionId,
		IsCorrupted: false,
		SampleNode:  sampleNode,
	}
}
