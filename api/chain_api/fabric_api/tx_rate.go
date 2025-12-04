package fabric_api

import (
	"fmt"
	"time"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
)

// StartTxRateTestCore 进行性能测试
func StartTxRateTestCore(coroutineCount int) error {
	// 创建合约
	contract, err := GetContract()
	if err != nil {
		return fmt.Errorf("get contract error: %w", err)
	}
	for i := 0; i < coroutineCount; i++ {
		rate_recorder.TxRateRecorderInstance.WaitGroup.Add(1)
		go func() {
			defer rate_recorder.TxRateRecorderInstance.WaitGroup.Done()
			currentCorountineTxCount := 0
			now := time.Now()
			for {
				select {
				case <-rate_recorder.TxRateRecorderInstance.StopQueue:
					return
				default:
					currentCorountineTxCount++
					assetId := fmt.Sprintf("asset-%d-%d-%d", i, now.Unix()*1e3+int64(now.Nanosecond())/1e6, currentCorountineTxCount)
					err := CreateAsset(contract, assetId)
					if err != nil {
						fmt.Printf("create asset error: %v\n", err)
						return
					}
					rate_recorder.TxRateRecorderInstance.TxCount.Add(1)
					//time.Sleep(time.Millisecond * 8)
				}
			}
		}()
	}
	return nil
}
