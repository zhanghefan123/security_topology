package chainmaker_api

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"sync/atomic"
	"time"
	"zhanghefan123/security_topology/api/chain_api/chainmaker_api/implementations"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
)

// StartTxRateTestAsyncCore 异步执行
func StartTxRateTestAsyncCore(trr *rate_recorder.TxRateRecorder, threadCount int) error {
	// 合约的名称
	contractName := fmt.Sprintf("fact%d", currentContractId)
	// 创建配置
	clientConfiguration := implementations.NewClientConfiguration(contractName)
	// 创建长安链的客户端
	chainMakerClient, err := implementations.CreateChainMakerClient(clientConfiguration)
	if err != nil {
		fmt.Printf("create chainmaker client failed, err:%v\n", err)
		return fmt.Errorf("cannot create chainmaker client: %w", err)
	}
	// 进行合约的创建
	err = implementations.CreateUpgradeUserContract(chainMakerClient, clientConfiguration, implementations.CreateContractOp)
	if err != nil {
		rate_recorder.TxRateRecorderInstance = nil
		fmt.Printf("create upgrade user contract failed, err:%v\n", err)
		return fmt.Errorf("cannot create user contract: %w", err)
	} else {
		currentContractId = currentContractId + 1
	}
	// 进行合约的调用
	count := 1                            // 序号
	var txCount atomic.Uint64             // 当前的合约执行次数
	var calcTpsDuration = time.Second * 1 // 计算的时间间隔

	// 启动多个执行线程
	for i := 0; i < threadCount; i++ { // 线程数量
		trr.WaitGroup.Add(1) // 正在执行的任务数 + 1
		go func() {
			defer trr.WaitGroup.Done()
		forLoop:
			for {
				select {
				case <-trr.StopQueue:
					break forLoop
				case trr.WorkerPool <- struct{}{}:
					go func() {
						defer func() {
							<-trr.WorkerPool
						}()
						var resp *common.TxResponse
						resp, err = implementations.InvokeContract(contractName, chainMakerClient, "save", true)
						if err != nil {
							fmt.Printf("invoke contract error: %v with response %v\n", err, resp)
						} else {
							txCount.Add(1)
						}
					}()
				}
			}
		}()
	}

	trr.WaitGroup.Add(1)
	go func(stopQueue chan struct{}) {
		defer trr.WaitGroup.Done()
	forLoop:
		for {
			select {
			case <-stopQueue:
				break forLoop
			default:
				txNum := txCount.Swap(0)
				tpsRate := float64(txNum) / calcTpsDuration.Seconds()
				if len(trr.RateList) == trr.FixedLength {
					trr.RateList = trr.RateList[1:]
					trr.RateList = append(trr.RateList, tpsRate)
					trr.TimeList = trr.TimeList[1:]
					trr.TimeList = append(trr.TimeList, count)
				} else {
					trr.RateList = append(trr.RateList, tpsRate)
					trr.TimeList = append(trr.TimeList, count)
				}
				trr.RateList = append(trr.RateList, tpsRate)
				trr.TimeList = append(trr.TimeList, count)
				count += 1
				time.Sleep(calcTpsDuration)
			}
		}
	}(trr.StopQueue)

	return nil
}
