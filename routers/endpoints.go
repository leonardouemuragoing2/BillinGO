package routers

import (
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GinRouter struct {
	Engine  *gin.Engine
	IRoutes gin.IRoutes
}

// Setup api endpoints for tests
func SetupEndpoints(db *gorm.DB) GinRouter {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	api := r.Use(limit.MaxAllowed(30))
	{
		api = GetEndpoints(api)
	}

	return GinRouter{r, api}
}

func GetEndpoints(api gin.IRoutes) gin.IRoutes {
	api.GET("/data", ListData)
	api.GET("/manager", ListManagerRRDData)

	return api
}
