package models

import "time"

type CancelingSituation struct {
	ID          int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Description string    `gorm:"column:description;not null" json:"description"`
	Data        int32     `gorm:"column:data;not null" json:"data"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
