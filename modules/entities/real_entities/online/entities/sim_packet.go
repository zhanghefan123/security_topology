package entities

type SimPacket struct {
	Path        *SimPath // 完整的端到端路径
	SessionId   string   // 这个 packet 沿着哪个 session 传递
	IsCorrupted bool     // 是否已经被篡改了
}

func CreatePacket(path *SimPath, sessionId string) *SimPacket {
	return &SimPacket{
		Path:        path,
		SessionId:   sessionId,
		IsCorrupted: false,
	}
}
