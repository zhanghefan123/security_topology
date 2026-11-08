package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/services/http_server/params/regenerate"
)

func RegenerateMultipathInformtion(c *gin.Context) {
	// 进行信息的重新的生成
	if topology.Instance != nil {
		c.JSON(http.StatusOK, gin.H{
			"messsage": fmt.Sprintf("topology regenerated"),
		})
	}

	// 进行信息的绑定
	regenerateMultipathInformation := &regenerate.RegenerateMultipathInformation{}
	err := c.BindJSON(regenerateMultipathInformation)
	if err != nil {
		fmt.Printf("regenerate multipath information failed\n")
	} else {
		fmt.Printf("regenerate multipath information succeed\n")
	}

	// 进行信息的重新的生成
	err = topology.Instance.RemoveAndRegenerateMultipath(regenerateMultipathInformation.MultipathType, regenerateMultipathInformation.MultipathFileName)
	if err != nil {
		fmt.Printf("regenerate multipath information failed\n")
	} else {
		fmt.Printf("regenerate multipath information succeed\n")
	}
}
