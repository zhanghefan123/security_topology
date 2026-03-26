package entities

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

type SimPathValidationRouter struct {
	*SimNodeBase
	SessionToFilterMapping map[string]*bloom.BloomFilter
	Weights                []float64
	ExploreProbabilities   []float64
	RectifiedGains         []float64
}

func NewSimPathValidationRouter(NodeName string, NodeIndex int) *SimPathValidationRouter {
	simNodeBase := CreateSimNodeBase(NodeName, NodeIndex)
	return &SimPathValidationRouter{
		SimNodeBase:            simNodeBase,
		SessionToFilterMapping: make(map[string]*bloom.BloomFilter),
		Weights:                make([]float64, 0),
		ExploreProbabilities:   make([]float64, 0),
		RectifiedGains:         make([]float64, 0),
	}
}

// ProcessPacket 进行数据包的处理
func (pathValidationRouter *SimPathValidationRouter) ProcessPacket(simPacket *SimPacket) (bool, error) {
	// find corresponding counter
	if bloomFilter, ok := pathValidationRouter.SessionToFilterMapping[simPacket.SessionId]; ok {
		var dropPacket bool
		if simPacket.IsCorrupted {
			dropPacket = true
		} else {
			// judge need to sample
			if simPacket.SamplePvRouter.NodeName == pathValidationRouter.NodeName {
				uuidBytes := []byte(simPacket.Uuid)
				bloomFilter.Add(uuidBytes)
			}
			dropPacket = false
		}
		return dropPacket, nil
	} else {
		return true, fmt.Errorf("process packet failed due to cannot find sessionid %s corresponding bloom filter", simPacket.SessionId)
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
