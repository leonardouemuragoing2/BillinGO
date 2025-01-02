package routers

import (
	"billingo/controllers"

	"github.com/gin-gonic/gin"
)

func ListManagerRRDData(c *gin.Context) {
	manager := c.MustGet("manager").(*controllers.Manager)

	// Retrieve all vmRRDData from the manager
	vmData := manager.GetAllVMData()

	// Return the data as JSON
	c.JSON(200, vmData)
}
