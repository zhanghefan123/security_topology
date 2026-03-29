package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimEndHost struct {
	*SimNodeBase
	SessionToCounterMapping map[string]*Counter // 对应于 per batch bloom filter 场景
	CurrentSelectedPath     *SimPath            // 当前源所选择的路径
	AckCounters             []int               // 记录 ack 计数
	Potential               float64             // 进行投影回正常平面用到的参数
}

func NewEndHost(NodeName string, NodeIndex int) *SimEndHost {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimEndHost{
		SimNodeBase:             simNodeBase,
		SessionToCounterMapping: make(map[string]*Counter),
		CurrentSelectedPath:     nil,
		AckCounters:             nil,
		Potential:               0,
	}
}

func (endHost *SimEndHost) SetCurrentEpochSelectedPath(currentEpochSelectedPath *SimPath) {
	endHost.CurrentSelectedPath = currentEpochSelectedPath
	endHost.AckCounters = make([]int, len(endHost.CurrentSelectedPath.NodeNameToIndexMapping))
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

func (endHost *SimEndHost) ProcessPacket(simPacket *SimPacket, simulationStrategy types.SimStrategy) (*SimPacket, error) {
	var ackPacket *SimPacket = nil
	if simPacket.Type == types.SimPacketType_DataPacket {
		if counter, ok := endHost.SessionToCounterMapping[simPacket.SessionId]; ok {
			if simPacket.Type == types.SimPacketType_DataPacket {
				if simulationStrategy == types.SimStrategy_PerBatchBloomFilter {
					if !simPacket.IsCorrupted {
						counter.ReceivedLegalPkts += 1
					}
					return nil, nil
				} else if simulationStrategy == types.SimStrategy_PerPacketAck {
					// judge need to sample
					sampleNodeName, _ := simPacket.SampleNode.GetSimNodeName()
					if sampleNodeName == endHost.NodeName {
						ackPacket = CreateSimPacket(types.SimPacketType_AckPacket, simPacket.SessionId, simPacket.SampleNode)
						return ackPacket, nil
					}
					return nil, nil
				} else {
					return nil, fmt.Errorf("unsupported packet type")
				}
			} else {
				return nil, fmt.Errorf("not supported packet type")
			}
		} else {
			return nil, fmt.Errorf("process packet failed due to cannot find sessionid %s corresponding counter", simPacket.SessionId)
		}
	} else if simPacket.Type == types.SimPacketType_AckPacket {
		if !simPacket.IsCorrupted {
			sampleNodeName, _ := simPacket.SampleNode.GetSimNodeName()
			endHost.AckCounters[endHost.CurrentSelectedPath.NodeNameToIndexMapping[sampleNodeName]] += 1
		}
		return nil, nil
	} else {
		return nil, fmt.Errorf("unsupported packet type")
	}
}
