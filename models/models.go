package models

import "time"

// BaseModel struct to be used in all models
type BaseModel struct {
	ID        uint      `gorm:"primary_key" json:"id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-21T16:33:51.147843-03:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-05-21T15:00:49.117789-03:00"`
}
