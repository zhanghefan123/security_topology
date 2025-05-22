package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
)

func InstallChannelAndChaincode(c *gin.Context) {
	// 1. 判断拓扑是否已经启动
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "topology instance is nil",
		})
		return
	}

	// 2. 判断是否是 fabric
	if topology.TopologyInstance.TopologyParams.BlockChainType != "fabric" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "blockchain type is not fabric",
		})
		return
	}

	// 3. 判断是否进行了链码的安装
	if topology.TopologyInstance.ChannelAndChainCodeInstalled {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "chaincode already installed",
		})
		return
	}

	// 4. 进行链码的安装
	testNetworkPath := configs.TopConfiguration.FabricConfig.FabricNetworkPath
	err := dir.WithContextManager(testNetworkPath, func() error {
		installChannelSh := fmt.Sprintf("./startInstallChannel.sh %d %d", len(topology.TopologyInstance.FabricOrdererNodes), len(topology.TopologyInstance.FabricPeerNodes))
		err := execute.Command("bash", []string{"-l", "-c", installChannelSh})
		if err != nil {
			return fmt.Errorf("start install channel failed: %w", err)
		}
		installChainCodeSh := fmt.Sprintf("./startInstallChaincode.sh %d %d", len(topology.TopologyInstance.FabricOrdererNodes), len(topology.TopologyInstance.FabricPeerNodes))
		err = execute.Command("bash", []string{"-l", "-c", installChainCodeSh})
		if err != nil {
			return fmt.Errorf("start install chaincode failed: %w", err)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("install channel and chaincode error: %s", err.Error()),
		})
		return
	}

	// 5. 返回正在测试的结果
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop",
	})
}
