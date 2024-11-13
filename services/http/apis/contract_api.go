package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/api/chainmaker_api"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
)

func CreateContract(c *gin.Context) {
	contractName := "fact"
	clientConfiguration := chainmaker_api.NewClientConfiguration(contractName)
	chainMakerClient, err := chainmaker_api.CreateChainMakerClient(clientConfiguration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not create chainmaker client",
		})
	}
	err = chainmaker_api.CreateUpgradeUserContract(chainMakerClient, clientConfiguration, chainmaker_api.CreateContractOp)
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

func StartTxRateTestRequest(c *gin.Context) {
	// 1. 判断是否拓扑已经启动
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}
	// 2. 判断是否存在共识节点
	if len(topology.TopologyInstance.ChainmakerNodes) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "chainmaker nodes is empty",
		})
		return
	}
	// 3. 判断是否已经启动了节点
	if chainmaker_api.TxRateRecorderInstance == nil { // 如果还没有启动 tps 测试
		chainmaker_api.TxRateRecorderInstance = chainmaker_api.NewTxRateRecorder()
		err := chainmaker_api.TxRateRecorderInstance.StartTxRateTest(topology.TopologyInstance.TopologyParams.ConsensusThreadCount)
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
			"time_list": chainmaker_api.TxRateRecorderInstance.TimeList,
			"rate_list": chainmaker_api.TxRateRecorderInstance.RateList,
		})
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
	// 2. 判断是否已经没有处在测试状态
	if chainmaker_api.TxRateRecorderInstance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "already not in a test state",
		})
		return
	}

	chainmaker_api.TxRateRecorderInstance.StopTxRateTest()
	chainmaker_api.TxRateRecorderInstance = nil

	// 3. 返回正在测试的结果
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop",
	})
}
