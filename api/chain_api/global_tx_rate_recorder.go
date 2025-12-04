package chain_api

import (
	"fmt"
	"os"
	"sync"
	"time"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/utils/file"
)

var GloablTxRateRecorderInstance *GlobalTxRateRecorder = nil

type TimestampAndRate struct {
	Timestamp time.Time
	Rate      float64
}

type GlobalTxRateRecorder struct {
	BlockChainType          types.ChainType
	TimestampAndRateListAll []*TimestampAndRate
	TickerChannel           chan struct{}
	StopQueue               chan struct{}
	WaitGroup               sync.WaitGroup
	FixedLength             int

	TimeList  []int
	RateList  []float64
	TimeCount int
}

func NewGlobalTxRateRecorder(blockChainType types.ChainType) *GlobalTxRateRecorder {
	return &GlobalTxRateRecorder{
		BlockChainType:          blockChainType,
		TimestampAndRateListAll: make([]*TimestampAndRate, 0),
		TickerChannel:           make(chan struct{}),
		StopQueue:               make(chan struct{}),
		WaitGroup:               sync.WaitGroup{},
		FixedLength:             10,

		TimeList:  make([]int, 0),
		RateList:  make([]float64, 0),
		TimeCount: 1,
	}
}

func StartGlobalTxRateRecorder(chainType types.ChainType) {
	if GloablTxRateRecorderInstance == nil {
		GloablTxRateRecorderInstance = NewGlobalTxRateRecorder(chainType)
	}
	GloablTxRateRecorderInstance.WaitGroup.Add(1)
	go func() {
		defer GloablTxRateRecorderInstance.WaitGroup.Done()
	forLoop:
		for {
			select {
			case <-GloablTxRateRecorderInstance.StopQueue:
				break forLoop
			case <-GloablTxRateRecorderInstance.TickerChannel:
				// 看当前是哪个链
				var txNum uint64
				var tpsRate float64
				if rate_recorder.TxRateRecorderInstance != nil {
					txNum = rate_recorder.TxRateRecorderInstance.TxCount.Swap(0)
					tpsRate = float64(txNum)
					// 向 TxRateRecorderInstance 之中进行添加
					AddTpsRate(tpsRate)
				} else {
					tpsRate = 0
				}
				timestampAndRate := &TimestampAndRate{
					Timestamp: time.Now(),
					Rate:      tpsRate,
				}
				//fmt.Println("tps rate: ", tpsRate);
				time.Now()
				GloablTxRateRecorderInstance.TimestampAndRateListAll = append(GloablTxRateRecorderInstance.TimestampAndRateListAll, timestampAndRate)
				AddTpsRateToGlobal(tpsRate)
			}
		}
	}()
}

func AddTpsRateToGlobal(tpsRate float64) {
	if GloablTxRateRecorderInstance != nil {
		if len(GloablTxRateRecorderInstance.RateList) == GloablTxRateRecorderInstance.FixedLength {
			GloablTxRateRecorderInstance.RateList = GloablTxRateRecorderInstance.RateList[1:]
			GloablTxRateRecorderInstance.RateList = append(GloablTxRateRecorderInstance.RateList, tpsRate)
			GloablTxRateRecorderInstance.TimeList = GloablTxRateRecorderInstance.TimeList[1:]
			GloablTxRateRecorderInstance.TimeList = append(GloablTxRateRecorderInstance.TimeList, GloablTxRateRecorderInstance.TimeCount)
		} else {
			GloablTxRateRecorderInstance.RateList = append(GloablTxRateRecorderInstance.RateList, tpsRate)
			GloablTxRateRecorderInstance.TimeList = append(GloablTxRateRecorderInstance.TimeList, GloablTxRateRecorderInstance.TimeCount)
		}
		GloablTxRateRecorderInstance.TimeCount += 1
	}
}

func StopGloablTxRateRecorder() error {
	if GloablTxRateRecorderInstance != nil {
		GloablTxRateRecorderInstance.StopQueue <- struct{}{}
		GloablTxRateRecorderInstance.WaitGroup.Wait()
		err := GloablTxRateRecorderInstance.WriteResultIntoFile()
		if err != nil {
			return fmt.Errorf("write result into file error: %v", err)
		}
		GloablTxRateRecorderInstance = nil
	}
	return nil
}

func (gtr *GlobalTxRateRecorder) WriteResultIntoFile() error {
	directory := fmt.Sprintf("./result/result%d", variables.UserSelectedExperimentNumber)
	// 进行相应的文件夹的创建
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdirall error: %v", err)
	}

	finalString := ""
	for index := 0; index < len(gtr.TimestampAndRateListAll); index++ {
		timestampAndRate := gtr.TimestampAndRateListAll[index]
		finalString += fmt.Sprintf("time: %v, value: %v\n", timestampAndRate.Timestamp.Format("15:04:05.000"), timestampAndRate.Rate)
	}
	// 将所有的序列放到一个文件之中
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s_result.txt", directory, gtr.BlockChainType.String()),
		finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}
	return nil
}
