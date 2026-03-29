package entities

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"

	"github.com/bits-and-blooms/bloom/v3"
)

type SimPathValidationRouter struct {
	*SimNodeBase
	SessionToFilterMapping map[string]*bloom.BloomFilter
	Weights                []float64
	ExploreProbabilities   []float64
	RectifiedGains         []float64
	Potential              float64
}

func NewSimPathValidationRouter(NodeName string, NodeIndex int) *SimPathValidationRouter {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimPathValidationRouter{
		SimNodeBase:            simNodeBase,
		SessionToFilterMapping: make(map[string]*bloom.BloomFilter),
		Weights:                make([]float64, 0),
		ExploreProbabilities:   make([]float64, 0),
		RectifiedGains:         make([]float64, 0),
		Potential:              0,
	}
}

// ProcessPacket 进行数据包的处理
func (pathValidationRouter *SimPathValidationRouter) ProcessPacket(simPacket *SimPacket, simulationStrategy types.SimStrategy) (bool, *SimPacket, error) {
	var ackPacket *SimPacket = nil
	var dropPacket = false
	// find corresponding counter
	if bloomFilter, ok := pathValidationRouter.SessionToFilterMapping[simPacket.SessionId]; ok {
		if simPacket.Type == types.SimPacketType_DataPacket {
			if simPacket.IsCorrupted {
				dropPacket = true
			} else {
				dropPacket = false
				if simulationStrategy == types.SimStrategy_PerBatchBloomFilter {
					// judge need to sample
					sampleNodeName, _ := simPacket.SampleNode.GetSimNodeName()
					if sampleNodeName == pathValidationRouter.NodeName {
						uuidBytes := []byte(simPacket.Uuid)
						bloomFilter.Add(uuidBytes)
					}
				} else if simulationStrategy == types.SimStrategy_PerPacketAck {
					if simPacket.Type == types.SimPacketType_DataPacket {
						// judge need to sample
						sampleNodeName, _ := simPacket.SampleNode.GetSimNodeName()
						if sampleNodeName == pathValidationRouter.NodeName {
							ackPacket = CreateSimPacket(types.SimPacketType_AckPacket, simPacket.SessionId, simPacket.SampleNode)
						}
					}
				} else {
					return true, nil, fmt.Errorf("unsupported ")
				}
			}
			return dropPacket, ackPacket, nil
		} else if simPacket.Type == types.SimPacketType_AckPacket { // 对于 ack packet 检测不了, 直接放行
			return false, nil, nil
		} else {
			return true, nil, fmt.Errorf("unsupported packet type")
		}

	} else {
		return true, nil, fmt.Errorf("process packet failed due to cannot find sessionid %s corresponding bloom filter", simPacket.SessionId)
	}
}

// EstablishSession 进行会话的建立
func (pathValidationRouter *SimPathValidationRouter) EstablishSession(sessionId string, M uint, K uint) error {
	if _, ok := pathValidationRouter.SessionToFilterMapping[sessionId]; !ok {
		pathValidationRouter.SessionToFilterMapping[sessionId] = bloom.New(M, K)
		return nil
	} else {
		return fmt.Errorf("cannot establish session %s since it has already existed", sessionId)
	}
}

// DestroySession 进行会话的拆除
func (pathValidationRouter *SimPathValidationRouter) DestroySession(sessionId string) error {
	if _, ok := pathValidationRouter.SessionToFilterMapping[sessionId]; ok {
		delete(pathValidationRouter.SessionToFilterMapping, sessionId)
		return nil
	} else {
		return fmt.Errorf("already been deleted")
	}
}

// RetrieveRecorder 获取各个 pvRouter 的布隆过滤器信息
func (pathValidationRouter *SimPathValidationRouter) RetrieveRecorder(sessionId string) (*SimRecorder, error) {
	if bloomFilter, ok := pathValidationRouter.SessionToFilterMapping[sessionId]; ok {
		oldBloomFilterCopy := bloomFilter.Copy()
		// clear the bit set
		bloomFilter.ClearAll()
		// create simRecorder
		simRecorder := &SimRecorder{
			RecorderType: types.SimRecorderType_BloomFilterRecorder,
			Counter:      nil,
			BloomFilter:  oldBloomFilterCopy,
		}
		return simRecorder, nil
	} else {
		return nil, fmt.Errorf("fail to retrieve counter of path validation router %s", pathValidationRouter.NodeName)
	}
}
