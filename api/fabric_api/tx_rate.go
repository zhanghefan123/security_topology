package fabric_api

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"sync"
	"sync/atomic"
	"time"
)

var (
	TxRateRecorderInstance *TxRateRecorder
)

type TxRateRecorder struct {
	TimeList    []int
	RateList    []float64
	fixedLength int
	stopQueue   chan struct{}
	waitGroup   *sync.WaitGroup
	done        chan struct{}
}

// NewTxRateRecorder 创建新的 TxRateRecorder
func NewTxRateRecorder() *TxRateRecorder {
	return &TxRateRecorder{
		TimeList:    make([]int, 0),
		RateList:    make([]float64, 0),
		fixedLength: 10,
		stopQueue:   make(chan struct{}),
		waitGroup:   &sync.WaitGroup{},
		done:        make(chan struct{}, 1),
	}
}

// StartTxRateTestCore 启动 Tx rate 测试的核心逻辑
func (trr *TxRateRecorder) StartTxRateTestCore(contract *client.Contract, coroutineCount int) {
	go func() {
		count := 1                            // 序号
		var txCount int64                     // 当前的合约执行次数
		var calcTpsDuration = time.Second * 1 // 计算的时间间隔
		for i := 0; i < coroutineCount; i++ {
			trr.waitGroup.Add(1)
			go func() {
				defer trr.waitGroup.Done()
				currentCorountineTxCount := 0
				now := time.Now()
				for {
					select {
					case <-trr.stopQueue:
						return
					default:
						txCount++
						currentCorountineTxCount++
						assetId := fmt.Sprintf("asset-%d-%d-%d", i, now.Unix()*1e3+int64(now.Nanosecond())/1e6, currentCorountineTxCount)
						err := CreateAsset(contract, assetId)
						if err != nil {
							fmt.Printf("create asset error: %v\n", err)
							return
						}
					}
				}
			}()
		}
		go func() {
			for {
				select {
				case <-trr.stopQueue:
					return
				default:
					// 计算 tps
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
					fmt.Println(trr.RateList)
					count += 1
					time.Sleep(calcTpsDuration)
				}
			}
		}()
		trr.waitGroup.Wait()
	}()
}

// StopTxRateTestCore 停止 Tx rate 的计算
func (trr *TxRateRecorder) StopTxRateTestCore() {
	close(trr.stopQueue)
	fmt.Println("stop tx rate test")
}

// StartTxRateTest 启动 Tx rate 测试
func (trr *TxRateRecorder) StartTxRateTest(coroutineCount int) error {
	contract, err := GetContract()
	if err != nil {
		return fmt.Errorf("get contract error: %w", err)
	}
	TxRateRecorderInstance = NewTxRateRecorder()
	TxRateRecorderInstance.StartTxRateTestCore(contract, coroutineCount)
	return nil
}
