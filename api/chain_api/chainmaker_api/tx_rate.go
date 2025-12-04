package chainmaker_api

import (
	"chainmaker.org/chainmaker/pb-go/v2/common"
	chainmaker_sdk_go "chainmaker.org/chainmaker/sdk-go/v2"
	"fmt"
	"zhanghefan123/security_topology/api/chain_api/chainmaker_api/implementations"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
)

var (
	currentContractId = 0
)

func CreateClientAndContract() (string, *chainmaker_sdk_go.ChainClient, error) {
	// 合约的名称
	contractName := fmt.Sprintf("fact%d", currentContractId)
	// 创建配置
	clientConfiguration := implementations.NewClientConfiguration(contractName)
	// 创建长安链的客户端
	chainMakerClient, err := implementations.CreateChainMakerClient(clientConfiguration)
	if err != nil {
		fmt.Printf("create chainmaker client failed, err:%v\n", err)
		return contractName, nil, fmt.Errorf("cannot create chainmaker client: %w", err)
	}
	// 进行合约的创建
	err = implementations.CreateUpgradeUserContract(chainMakerClient, clientConfiguration, implementations.CreateContractOp)
	if err != nil {
		fmt.Printf("create upgrade user contract failed, err:%v\n", err)
		return "", nil, fmt.Errorf("cannot create user contract: %w", err)
	} else {
		currentContractId = currentContractId + 1
	}
	return contractName, chainMakerClient, nil
}

// StartTxRateTestCore 进行 tx 速率的测试
func StartTxRateTestCore(threadCount int) error {
	contractName, chainMakerClient, err := CreateClientAndContract()
	if err != nil {
		return fmt.Errorf("start tx rate test core error: %v", err)
	}

	var resp *common.TxResponse        // 合约执行的响应结果
	for i := 0; i < threadCount; i++ { // 线程数量
		rate_recorder.TxRateRecorderInstance.WaitGroup.Add(1) // 正在执行的任务数 + 1
		go func() {
			defer rate_recorder.TxRateRecorderInstance.WaitGroup.Done()
		forLoop:
			for {
				select {
				case <-rate_recorder.TxRateRecorderInstance.StopQueue:
					break forLoop
				default:
					resp, err = implementations.InvokeContract(contractName, chainMakerClient, "save", true)
					if err != nil {
						fmt.Printf("%s\ninvoke contract resp: %+v\n", err, resp)
					} else {
						rate_recorder.TxRateRecorderInstance.TxCount.Add(1)
					}
				}
			}
		}()
	}

	return nil
}
