package routers

import (
	"billingo/filters"
	"billingo/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListServers list all data
// @Summary List all data in dB
// @Produce json
// @Tags Servers
// @Success 200 {object} object{items=[]models.Data}
// @Failure 400,500 {object} object{error=string}
// @Router /data [get]
func ListData(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var dataFilter filters.DataFilter
	if err := c.ShouldBindQuery(&dataFilter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var data []models.Data
	query := db.Model(&models.Data{}) //.Where("is_active", true)
	dataFilter.Filter(query, nil).Order("time DESC").Find(&data)

	// Set response
	c.JSON(http.StatusOK, gin.H{"items": data})
}
