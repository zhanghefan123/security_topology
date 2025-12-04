package chain_api

import (
	"fmt"
	"zhanghefan123/security_topology/api/chain_api/chainmaker_api"
	"zhanghefan123/security_topology/api/chain_api/fabric_api"
	"zhanghefan123/security_topology/api/chain_api/fisco_bcos_api"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
	"zhanghefan123/security_topology/modules/entities/types"
)

func StartTxRateTest(coroutineCount int, chainType types.ChainType) error {
	if rate_recorder.TxRateRecorderInstance == nil {
		rate_recorder.TxRateRecorderInstance = rate_recorder.NewTxRateRecorder()
	}
	if chainType == types.ChainType_ChainMaker {
		err := chainmaker_api.StartTxRateTestCore(coroutineCount)
		if err != nil {
			return fmt.Errorf("start tx rate test core error: %w", err)
		} else {
			return nil
		}
	} else if chainType == types.ChainType_HyperledgerFabric {
		err := fabric_api.StartTxRateTestCore(coroutineCount)
		if err != nil {
			return fmt.Errorf("start tx rate test core error: %w", err)
		} else {
			return nil
		}
	} else if chainType == types.ChainType_FiscoBcos {
		err := fisco_bcos_api.StartTxRateTestCore(coroutineCount)
		if err != nil {
			return fmt.Errorf("start tx rate test core error: %w", err)
		} else {
			return nil
		}
	} else {
		panic("not supported chain type")
	}
}

// StopTxRateTest 停止 tx rate 速率的测试
func StopTxRateTest() {
	if rate_recorder.TxRateRecorderInstance == nil {
		return
	} else {
		close(rate_recorder.TxRateRecorderInstance.StopQueue)
		rate_recorder.TxRateRecorderInstance.WaitGroup.Wait()
		fmt.Println("stop tx rate test")
	}
}

func AddTpsRate(tpsRate float64) {
	if rate_recorder.TxRateRecorderInstance != nil {
		if len(rate_recorder.TxRateRecorderInstance.RateList) == rate_recorder.TxRateRecorderInstance.FixedLength {
			rate_recorder.TxRateRecorderInstance.RateList = rate_recorder.TxRateRecorderInstance.RateList[1:]
			rate_recorder.TxRateRecorderInstance.RateList = append(rate_recorder.TxRateRecorderInstance.RateList, tpsRate)
			rate_recorder.TxRateRecorderInstance.TimeList = rate_recorder.TxRateRecorderInstance.TimeList[1:]
			rate_recorder.TxRateRecorderInstance.TimeList = append(rate_recorder.TxRateRecorderInstance.TimeList, rate_recorder.TxRateRecorderInstance.TimeCount)
		} else {
			rate_recorder.TxRateRecorderInstance.RateList = append(rate_recorder.TxRateRecorderInstance.RateList, tpsRate)
			rate_recorder.TxRateRecorderInstance.TimeList = append(rate_recorder.TxRateRecorderInstance.TimeList, rate_recorder.TxRateRecorderInstance.TimeCount)
		}
		rate_recorder.TxRateRecorderInstance.TimeCount += 1
	}
}
