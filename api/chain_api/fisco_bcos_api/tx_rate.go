package fisco_bcos_api

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"zhanghefan123/security_topology/api/chain_api/fisco_bcos_api/hello"
)

var (
	TxRateRecorderInstance *TxRateRecorder
)

type TxRateRecorder struct {
	TimeList    []int     // 存储固定长度的时间序列
	RateList    []float64 // 存储固定长度的速率
	fixedLength int       // 固定的长度
	TimeListAll []int     // 存储所有的时间序列
	RateListAll []float64 // 存储所有的速率序列
	stopQueue   chan struct{}
	waitGroup   *sync.WaitGroup
	done        chan struct{}
}

func NewTxRateRecorder() *TxRateRecorder {
	return &TxRateRecorder{
		TimeList:    make([]int, 0),
		RateList:    make([]float64, 0),
		fixedLength: 10,
		TimeListAll: make([]int, 0),
		RateListAll: make([]float64, 0),
		stopQueue:   make(chan struct{}),
		waitGroup:   &sync.WaitGroup{},
		done:        make(chan struct{}, 1),
	}
}

// TestNetworkLatency 测试网络延迟
func (trr *TxRateRecorder) TestNetworkLatency(helloWorldSession *hello.HelloWorldSession) {
	fmt.Println("=== 网络延迟测试开始 ===")
	for i := 0; i < 5; i++ {
		startTime := time.Now()
		_, err := helloWorldSession.Get() // 使用只读操作测试网络延迟
		timeDuration := time.Since(startTime)
		if err != nil {
			fmt.Printf("网络测试失败: %v\n", err)
			return
		}
		fmt.Printf("网络延迟测试 %d: %v\n", i+1, timeDuration)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("=== 网络延迟测试结束 ===")
}

func (trr *TxRateRecorder) StartTxRateTestCore(helloWorldSession *hello.HelloWorldSession, coroutineCount int) {
	// 先进行网络延迟测试
	trr.TestNetworkLatency(helloWorldSession)
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
						currentCorountineTxCount++
						txCount++
						assetId := fmt.Sprintf("asset-%d-%d-%d", i, now.Unix()*1e3+int64(now.Nanosecond())/1e6, currentCorountineTxCount)

						// 应该是内部的 batch 时间特别的长 --> 接近1s, 导致同步的机制无法及时进行处理
						_, err, _, _ := helloWorldSession.Set(assetId)
						//_, err := helloWorldSession.AsyncSet(func(receipt *types.Receipt, err error) {
						//	// 这个函数是设置的回调函数
						//	if err != nil {
						//		fmt.Printf("async set finish error: %v", err)
						//	} else {
						//		txCount++
						//	}
						//}, assetId)

						if err != nil {
							fmt.Printf("async set start error: %v\n", err)
							return
						}
						time.Sleep(time.Millisecond * 1)
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
					trr.RateListAll = append(trr.RateListAll, tpsRate)
					trr.TimeListAll = append(trr.TimeListAll, count)
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
	helloWorldSession := CreateSessionAndDeploySmartContract()
	TxRateRecorderInstance = NewTxRateRecorder()
	TxRateRecorderInstance.StartTxRateTestCore(helloWorldSession, coroutineCount)
	return nil
}
