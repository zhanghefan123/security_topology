package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/api/chain_api"
	"zhanghefan123/security_topology/api/chain_api/rate_recorder"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/modules/entities/types"
)

func GetTxRateTestRequest(c *gin.Context) {
	// 1. 判断是否拓扑已经启动
	if topology.Instance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 判断是否是 NonChainType
	if topology.Instance.TopologyParams.BlockChainType == types.ChainType_NonChain {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "no blockchain",
		})
		return
	}

	// 3. 直接进行返回
	if chain_api.GloablTxRateRecorderInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"time_list": chain_api.GloablTxRateRecorderInstance.TimeList,
			"rate_list": chain_api.GloablTxRateRecorderInstance.RateList,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"time_list": make([]int, 0),
			"rate_list": make([]float64, 0),
		})
	}
}

// StartTxRateTestRequest 是web request对应的回调函数
func StartTxRateTestRequest(c *gin.Context) {
	// 1. 判断是否拓扑已经启动
	if topology.Instance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 判断是否是 NonChainType
	if topology.Instance.TopologyParams.BlockChainType == types.ChainType_NonChain {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "no blockchain",
		})
		return
	}

	// 2. 获取携程数量
	coroutineCount := topology.Instance.TopologyParams.ConsensusThreadCount

	// 3. 进行启动
	if rate_recorder.TxRateRecorderInstance == nil {
		// 3.1 如果还没有启动 tps 测试
		err := chain_api.StartTxRateTest(coroutineCount, topology.Instance.TopologyParams.BlockChainType)
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
		// 3.2 如果已经启动了 tps 测试 -> 直接进行结果返回
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"time_list": rate_recorder.TxRateRecorderInstance.TimeList,
			"rate_list": rate_recorder.TxRateRecorderInstance.RateList,
		})
	}
}

func StopTxRateTestRequest(c *gin.Context) {
	// 1. 判断拓扑是否已经启动
	if topology.Instance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 如果不是区块链类型
	if topology.Instance.TopologyParams.BlockChainType == types.ChainType_NonChain {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "no blockchain",
		})
		return
	}

	// 3 判断是否已经没有处在测试状态
	if rate_recorder.TxRateRecorderInstance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "already not in a test state",
		})
		return
	}
	chain_api.StopTxRateTest()
	rate_recorder.TxRateRecorderInstance = nil

	// 3. 返回正在测试的结果
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop",
	})
}
