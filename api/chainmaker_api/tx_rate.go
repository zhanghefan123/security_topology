package chainmaker_api

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	currentContractId                      = 0
	TxRateRecorderInstance *TxRateRecorder = nil
)

type TxRateRecorder struct {
	TimeList    []int
	RateList    []float64
	fixedLength int
	stopQueue   chan struct{}
	waitGroup   *sync.WaitGroup
}

// NewTxRateRecorder 创建新的 TxRateRecorder
func NewTxRateRecorder() *TxRateRecorder {
	return &TxRateRecorder{
		TimeList:    make([]int, 0),
		RateList:    make([]float64, 0),
		fixedLength: 10,
		stopQueue:   make(chan struct{}),
		waitGroup:   &sync.WaitGroup{},
	}
}

// StartTxRateTest 进行 tx 速率的测试
func (trr *TxRateRecorder) StartTxRateTest(threadCount int) error {
	// 合约的名称
	contractName := fmt.Sprintf("fact%d", currentContractId)
	// 创建配置
	clientConfiguration := NewClientConfiguration(contractName)
	// 创建长安链的客户端
	chainMakerClient, err := CreateChainMakerClient(clientConfiguration)
	if err != nil {
		return fmt.Errorf("cannot create chainmaker client: %w", err)
	}
	// 进行合约的创建
	err = CreateUpgradeUserContract(chainMakerClient, clientConfiguration, CreateContractOp)
	if err != nil {
		TxRateRecorderInstance = nil
		return fmt.Errorf("cannot create user contract: %w", err)
	} else {
		currentContractId = currentContractId + 1
	}
	// 进行合约的调用
	count := 1                            // 序号
	var txCount int64                     // 当前的合约执行次数
	var calcTpsDuration = time.Second * 1 // 计算的时间间隔
	var resp *common.TxResponse           // 合约执行的响应结果
	for i := 0; i < threadCount; i++ {    // 线程数量
		trr.waitGroup.Add(1) // 正在执行的任务数 + 1
		go func(stopQueue chan struct{}) {
			defer trr.waitGroup.Done()
		forLoop:
			for {
				select {
				case <-stopQueue:
					break forLoop
				default:
					resp, err = invokeContract(contractName, chainMakerClient, "save", true)
					if err != nil {
						fmt.Printf("%s\ninvoke contract resp: %+v\n", err, resp)
					} else {
						atomic.AddInt64(&txCount, 1)
					}
				}
			}
		}(trr.stopQueue)
	}
	go func(stopQueue chan struct{}) {
	forLoop:
		for {
			select {
			case <-stopQueue:
				break forLoop
			default:
				txNum := atomic.SwapInt64(&txCount, 0)
				tpsRate := float64(txNum) / calcTpsDuration.Seconds()
				if len(trr.RateList) == trr.fixedLength {
					trr.RateList = trr.RateList[1:]
					trr.RateList = append(trr.RateList, tpsRate)
					trr.TimeList = trr.TimeList[1:]
					trr.TimeList = append(trr.TimeList, count)
				} else {
					trr.RateList = append(trr.RateList, tpsRate)
					trr.TimeList = append(trr.TimeList, count)
				}
				count += 1
				time.Sleep(calcTpsDuration)
			}
		}
	}(trr.stopQueue)

	return nil
}

// StopTxRateTest 停止 tx rate 速率的测试
func (trr *TxRateRecorder) StopTxRateTest() {
	close(trr.stopQueue)
	trr.waitGroup.Wait()
}
