package entities

import "chainmaker.org/chainmaker/common/v2/random/uuid"

type SimPacket struct {
	Path           *SimPath                 // 完整的端到端路径
	Uuid           string                   // 数据包的唯一的标识
	SessionId      string                   // 这个 packet 沿着哪个 session 传递
	IsCorrupted    bool                     // 是否已经被篡改了
	IsDropped      bool                     // 是否已经被丢弃
	SamplePvRouter *SimPathValidationRouter // 需要采样的 pv router
}

// CreateSimPacket 进行模拟数据包的创建
func CreateSimPacket(path *SimPath, sessionId string, samplePvRouter *SimPathValidationRouter) *SimPacket {
	return &SimPacket{
		Path:           path,
		Uuid:           uuid.GetUUID(),
		SessionId:      sessionId,
		IsCorrupted:    false,
		IsDropped:      false,
		SamplePvRouter: samplePvRouter,
	}
}
