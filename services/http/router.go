package http

import (
	"github.com/gin-gonic/gin"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	HttpServiceLogger = logger.GetLogger(logger.ModuleHttpService)
)

var postRoutes = map[string]gin.HandlerFunc{
	"/startConstellation": StartConstellation,
	"/stopConstellation":  StopConstellation,
	"/startTopology":      StartTopology,
	"/stopTopology":       StopTopology,
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
