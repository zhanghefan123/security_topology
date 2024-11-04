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
	contractName := fmt.Sprintf("fact%d", currentContractId)
	clientConfiguration := NewClientConfiguration(contractName)
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
	count := 1
	var txCount int64
	var calcTpsDuration = time.Second * 1
	var resp *common.TxResponse
	for i := 0; i < threadCount; i++ {
		trr.waitGroup.Add(1)
		go func(stopQueue chan struct{}) {
			defer trr.waitGroup.Done()
		forLoop:
			for {
				select {
				case <-stopQueue:
					break forLoop
				default:
					resp, err = testUserContractClaimInvoke(contractName, chainMakerClient, "save", true)
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
