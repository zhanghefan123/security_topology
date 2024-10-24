package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/api/chainmaker_api"
)

func CreateContract(c *gin.Context) {
	err := chainmaker_api.CreateUpgradeUserContract(chainmaker_api.CreateContractOp, "fact")
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
