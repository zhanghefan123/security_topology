package fisco_bcos_api

import (
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"sync/atomic"
	"time"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
)

func StartTxRateTestCoreAsync(coroutineCount int) error {
	helloWorldSession, err := CreateSessionAndDeploySmartContract()
	if err != nil {
		return fmt.Errorf("create session and deploy smart contract")
	}
	var txCount atomic.Uint64 // 当前的合约执行次数
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

					// 应该是内部的 batch 时间特别的长 --> 接近1s, 导致同步的机制无法及时进行处理
					_, err := helloWorldSession.AsyncSet(func(receipt *types.Receipt, err error) {
						// 这个函数是设置的回调函数
						if err != nil {
							fmt.Printf("async set finish error: %v", err)
						} else {
							txCount.Add(1)
						}
					}, assetId)

					if err != nil {
						fmt.Printf("async set start error: %v\n", err)
						return
					}
					time.Sleep(time.Millisecond * 1)
				}
			}
		}()
	}
	return nil
}

func StartTxRateTestCore(coroutineCount int) error {
	helloWorldSession, err := CreateSessionAndDeploySmartContract()
	if err != nil {
		return fmt.Errorf("create session and deploy smart contract")
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
					// 应该是内部的 batch 时间特别的长 --> 接近1s, 导致同步的机制无法及时进行处理
					_, err, _, _ := helloWorldSession.Set(assetId)
					if err != nil {
						fmt.Printf("async set start error: %v\n", err)
						return
					}
					rate_recorder.TxRateRecorderInstance.TxCount.Add(1)
					time.Sleep(time.Millisecond * 50)
				}
			}
		}()
	}
	return nil
}
