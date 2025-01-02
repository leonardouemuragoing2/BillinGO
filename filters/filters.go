package filters

import (
	"billingo/models"
	"time"

	"gorm.io/gorm"
)

type DataFilter struct {
	VMIDs           []uint     `form:"vmids" json:"vmids"`
	InitialDatetime *time.Time `form:"initial_datetime" json:"initial_datetime"`
	FinalDatetime   *time.Time `form:"final_datetime" json:"final_datetime"`
}

func (filter *DataFilter) Filter(query *gorm.DB, VMIDs *uint) *gorm.DB {
	if len(filter.VMIDs) > 0 {
		query = query.Model(models.Data{}).Where("vm_id IN ?", []int{int(*VMIDs)}).Order("time DESC")
	}
	var emptytimedate *time.Time

	if (filter.InitialDatetime != emptytimedate) && (filter.FinalDatetime != emptytimedate) {
		query = query.Where("(start_time <= ? and finish_time >= ?) OR (start_time >= ? and start_time <= ?) OR (start_time <= ? and finish_time IS NULL) OR (start_time >= ? and finish_time >= ? and start_time <= ?) OR (start_time <= ? and finish_time <= ? and finish_time >= ?) ", filter.InitialDatetime, filter.FinalDatetime, filter.InitialDatetime, filter.FinalDatetime, filter.InitialDatetime, filter.InitialDatetime, filter.FinalDatetime, filter.FinalDatetime, filter.InitialDatetime, filter.FinalDatetime, filter.InitialDatetime)
	}
	return query
}
