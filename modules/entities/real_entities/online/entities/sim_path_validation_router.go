package entities

import (
	"fmt"
)

type Counter struct {
	IllegalPackets int
	LegalPackets   int
}

type SimPathValidationRouter struct {
	*SimNodeBase
	SessionToCounterMapping map[string]*Counter
}

func NewSimPathValidationRouter(NodeName string, NodeIndex int) *SimPathValidationRouter {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimPathValidationRouter{
		SimNodeBase:             simNodeBase,
		SessionToCounterMapping: make(map[string]*Counter),
	}
}

// ProcessPacket 进行数据包的处理
func (pathValidationRouter *SimPathValidationRouter) ProcessPacket(pkt *SimPacket) (bool, error) {
	// find corresponding counter
	if counter, ok := pathValidationRouter.SessionToCounterMapping[pkt.SessionId]; ok {
		var dropPacket bool
		if pkt.IsCorrupted {
			counter.IllegalPackets += 1
			dropPacket = true
		} else {
			counter.LegalPackets += 1
			dropPacket = false
		}
		return dropPacket, nil
	} else {
		return true, fmt.Errorf("process packet failed due to cannot find sessionid %s corresponding counter", pkt.SessionId)
	}
}

// EstablishSession 进行会话的建立
func (pathValidationRouter *SimPathValidationRouter) EstablishSession(sessionId string) error {
	if _, ok := pathValidationRouter.SessionToCounterMapping[sessionId]; !ok {
		pathValidationRouter.SessionToCounterMapping[sessionId] = &Counter{
			IllegalPackets: 0,
			LegalPackets:   0,
		}
		return nil
	} else {
		return fmt.Errorf("cannot establish session %s since it has already existed", sessionId)
	}
}

// DestroySession 进行会话的拆除
func (pathValidationRouter *SimPathValidationRouter) DestroySession(sessionId string) error {
	if _, ok := pathValidationRouter.SessionToCounterMapping[sessionId]; ok {
		delete(pathValidationRouter.SessionToCounterMapping, sessionId)
		return nil
	} else {
		return fmt.Errorf("already been deleted")
	}
}

// RetrieveCounter 获取 counter 信息
func (pathValidationRouter *SimPathValidationRouter) RetrieveCounter(sessionId string) (*Counter, error) {
	if counter, ok := pathValidationRouter.SessionToCounterMapping[sessionId]; ok {
		oldCounter := &Counter{
			IllegalPackets: counter.IllegalPackets,
			LegalPackets:   counter.LegalPackets,
		}
		counter.IllegalPackets = 0
		counter.LegalPackets = 0
		return oldCounter, nil
	} else {
		return nil, fmt.Errorf("fail to retrieve counter of path validation router %s", pathValidationRouter.NodeName)
	}
}
