package entities

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
)

// Counter 计数器
type Counter struct {
	ReceivedLegalPkts int
}

// SimRecorder 目的节点返回的是 counter, 其他节点返回的是 bloom filter
type SimRecorder struct {
	RecorderType types.SimRecorderType
	BloomFilter  *bloom.BloomFilter
	Counter      *Counter
}

// GetValidatedPacketsCount 获取合法的数据包的数量
func (simRecorder *SimRecorder) GetValidatedPacketsCount() (int, error) {
	if simRecorder.RecorderType == types.SimRecorderType_CountRecorder {
		return simRecorder.Counter.ReceivedLegalPkts, nil
	} else if simRecorder.RecorderType == types.SimRecorderType_BloomFilterRecorder {
		return int(simRecorder.BloomFilter.ApproximatedSize()), nil
	} else {
		return -1, fmt.Errorf("unsupported sim recorder type")
	}
}
