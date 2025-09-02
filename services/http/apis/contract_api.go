package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	chainmaker_api2 "zhanghefan123/security_topology/api/chain_api/chainmaker_api"
	"zhanghefan123/security_topology/api/chain_api/fabric_api"
	"zhanghefan123/security_topology/api/chain_api/fisco_bcos_api"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
)

func CreateContract(c *gin.Context) {
	contractName := "fact"
	clientConfiguration := chainmaker_api2.NewClientConfiguration(contractName)
	chainMakerClient, err := chainmaker_api2.CreateChainMakerClient(clientConfiguration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not create chainmaker client",
		})
	}
	err = chainmaker_api2.CreateUpgradeUserContract(chainMakerClient, clientConfiguration, chainmaker_api2.CreateContractOp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("create user contract error: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

// StartTxRateTestRequest 是web request对应的回调函数
func StartTxRateTestRequest(c *gin.Context) {
	// 1. 判断是否拓扑已经启动
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 判断区块链的类型
	// 2.1 如果是长安链
	if topology.TopologyInstance.TopologyParams.BlockChainType == "长安链" {
		// 2.1.1 判断是否存在长安链共识节点
		if !topology.TopologyInstance.ChainMakerEnabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "chainmaker nodes is empty",
			})
			return
		}
		// 2.1.2 判断是否已经启动了tps测试
		if chainmaker_api2.TxRateRecorderInstance == nil { // 如果还没有启动 tps 测试
			chainmaker_api2.TxRateRecorderInstance = chainmaker_api2.NewTxRateRecorder()
			err := chainmaker_api2.TxRateRecorderInstance.StartTxRateTest(topology.TopologyInstance.TopologyParams.ConsensusThreadCount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": fmt.Sprintf("start tx rate test error: %s", err.Error()),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": make([]int, 0),
				"rate_list": make([]float64, 0),
			})
		} else { // 如果已经启动了 tps 测试
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": chainmaker_api2.TxRateRecorderInstance.TimeList,
				"rate_list": chainmaker_api2.TxRateRecorderInstance.RateList,
			})
		}
	} else if topology.TopologyInstance.TopologyParams.BlockChainType == "fabric" { // 2.2 如果是 fabric 链
		// 2.2.1 判断是否存在 fabric 共识节点
		if !topology.TopologyInstance.FabricEnabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "fabric nodes is empty",
			})
			return
		}
		// 2.2.2 判断是否已经启动了 tps 测试
		if fabric_api.TxRateRecorderInstance == nil { // 如果还没有启动 tps 测试
			fabric_api.TxRateRecorderInstance = fabric_api.NewTxRateRecorder()
			err := fabric_api.TxRateRecorderInstance.StartTxRateTest(topology.TopologyInstance.TopologyParams.ConsensusThreadCount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": fmt.Sprintf("start tx rate test error: %s", err.Error()),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": make([]int, 0),
				"rate_list": make([]float64, 0),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": fabric_api.TxRateRecorderInstance.TimeList,
				"rate_list": fabric_api.TxRateRecorderInstance.RateList,
			})
		}
	} else if topology.TopologyInstance.TopologyParams.BlockChainType == "fisco-bcos" {
		// 2.3.1 判断是否存在 fisco bcos 共识节点
		if !topology.TopologyInstance.FiscoBcosEnabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "fisco bcos nodes is empty",
			})
			return
		}
		// 2.3.2 判断是否已经启动了 tps 测试
		if fisco_bcos_api.TxRateRecorderInstance == nil { // 如果还没有启动 tps 测试
			fisco_bcos_api.TxRateRecorderInstance = fisco_bcos_api.NewTxRateRecorder()
			err := fisco_bcos_api.TxRateRecorderInstance.StartTxRateTest(topology.TopologyInstance.TopologyParams.ConsensusThreadCount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": fmt.Sprintf("start tx rate test error: %s", err.Error()),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": make([]int, 0),
				"rate_list": make([]float64, 0),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"time_list": fisco_bcos_api.TxRateRecorderInstance.TimeList,
				"rate_list": fisco_bcos_api.TxRateRecorderInstance.RateList,
			})
		}
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "blockchain type is not selected",
		})
		return
	}
}

func StopTxRateTestRequest(c *gin.Context) {
	// 1. 判断拓扑是否已经启动
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 判断联盟链类型
	if topology.TopologyInstance.TopologyParams.BlockChainType == "长安链" {
		// 2.1 判断是否已经没有处在测试状态
		if chainmaker_api2.TxRateRecorderInstance == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "already not in a test state",
			})
			return
		}

		chainmaker_api2.TxRateRecorderInstance.StopTxRateTest()
		err := chainmaker_api2.TxRateRecorderInstance.WriteResultIntoFile()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "write result into file error",
			})
		}
		chainmaker_api2.TxRateRecorderInstance = nil
	} else if topology.TopologyInstance.TopologyParams.BlockChainType == "fabric" {
		// 2.2 判断是否已经没有处在测试状态
		if fabric_api.TxRateRecorderInstance == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "already not in a test state",
			})
			return
		}
		fabric_api.TxRateRecorderInstance.StopTxRateTestCore()
		err := fabric_api.TxRateRecorderInstance.WriteResultIntoFile()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "write result into file error",
			})
			fmt.Printf("write result into file error: %v", err)
			return
		}
		fabric_api.TxRateRecorderInstance = nil
	} else if topology.TopologyInstance.TopologyParams.BlockChainType == "fisco-bcos" {
		// 2.3 判断是否已经没有处在测试状态
		if fisco_bcos_api.TxRateRecorderInstance == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "already not in a test state",
			})
			return
		}
		fisco_bcos_api.TxRateRecorderInstance.StopTxRateTestCore()
		fisco_bcos_api.TxRateRecorderInstance = nil
	} else {
		fmt.Println("non blockchain nodes, cannot start contract test")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "already not in a test state",
		})
		return
	}

	// 3. 返回正在测试的结果
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop",
	})
}
