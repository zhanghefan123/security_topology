package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimEndHost struct {
	*SimNodeBase
	SessionToCounterMapping map[string]*Counter
}

func NewEndHost(NodeName string, NodeIndex int) *SimEndHost {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimEndHost{
		SimNodeBase:             simNodeBase,
		SessionToCounterMapping: make(map[string]*Counter),
	}
}

func (endHost *SimEndHost) EstablishSession(sessionId string) error {
	if _, ok := endHost.SessionToCounterMapping[sessionId]; !ok {
		endHost.SessionToCounterMapping[sessionId] = &Counter{
			ReceivedLegalPkts: 0,
		}
		return nil
	} else {
		return fmt.Errorf("cannot establish session %s since it has already existed", sessionId)
	}
}

func (endHost *SimEndHost) DestroySession(sessionId string) error {
	if _, ok := endHost.SessionToCounterMapping[sessionId]; ok {
		delete(endHost.SessionToCounterMapping, sessionId)
		return nil
	} else {
		return fmt.Errorf("already been deleted")
	}
}

func (endHost *SimEndHost) RetrieveRecorder(sessionId string) (*SimRecorder, error) {
	if counter, ok := endHost.SessionToCounterMapping[sessionId]; ok {
		simRecorder := &SimRecorder{
			RecorderType: types.SimRecorderType_CountRecorder,
			BloomFilter:  nil,
			Counter: &Counter{
				ReceivedLegalPkts: counter.ReceivedLegalPkts,
			},
		}
		counter.ReceivedLegalPkts = 0
		return simRecorder, nil
	} else {
		return nil, fmt.Errorf("cannot retrieve recorder")
	}
}

func (endHost *SimEndHost) ProcessPacket(simPacket *SimPacket) error {
	if counter, ok := endHost.SessionToCounterMapping[simPacket.SessionId]; ok {
		if !simPacket.IsCorrupted {
			counter.ReceivedLegalPkts += 1
		}
		return nil
	} else {
		return fmt.Errorf("process packet failed due to cannot find sessionid %s corresponding counter", simPacket.SessionId)
	}
}
