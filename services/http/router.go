package http

import (
	"github.com/gin-gonic/gin"
	"zhanghefan123/security_topology/services/http/apis"
)

var postRoutes = map[string]gin.HandlerFunc{
	"/createContract":  apis.CreateContract,         // contract 相关
	"/startTxRateTest": apis.StartTxRateTestRequest, // contract 相关
	"/stopTxRateTest":  apis.StopTxRateTestRequest,  // contract 相关

	"/installChannelAndChaincode": apis.InstallChannelAndChaincode, // fabric 相关

	"/getConstellationState": apis.GetConstellationState, // (constellation 相关)
	"/getInstancePositions":  apis.GetInstancesPositions, // (constellation 相关)
	"/startConstellation":    apis.StartConstellation,    // (constellation 相关)
	"/stopConstellation":     apis.StopConstellation,     // (constellation 相关)
	"/changeTimeStep":        apis.ChangeTimeStamp,       // (constellation 相关)

	"/getTopologyState":       apis.GetTopologyState,       // (topology 相关)
	"/startTopology":          apis.StartTopology,          // (topology 相关)
	"/stopTopology":           apis.StopTopology,           // (topology 相关)
	"/saveTopology":           apis.SaveTopology,           // (topology 相关)
	"/getTopologyDescription": apis.GetTopologyDescription, // (topology 相关)
	"/changeStartDefence":     apis.ChangeStartDefence,     // (topology 相关)

	"/startWebShell": apis.StartWebShell, // webshell 相关
	"/stopWebShell":  apis.StopWebShell,  // webshell 相关

	"/startCaptureInterfaceRate": apis.StartCaptureInstancePerformance, // rate 相关
	"/stopCaptureInterfaceRate":  apis.StopCaptureInstancePerformance,  // rate 相关
}

// CORSMiddleware 中间件处理跨域问题
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// InitRouter 进行 gin 引擎的创建
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	for route, callback := range postRoutes {
		r.POST(route, callback)
	}
	return r
}
