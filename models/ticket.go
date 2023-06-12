package models

import "time"

type Ticket struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID       int32     `gorm:"column:u_id;not null" json:"u_id"`
	PID       int32     `gorm:"column:p_id;not null" json:"p_id"`
	FID       int32     `gorm:"column:f_id;not null" json:"f_id"`
	Status    string    `gorm:"column:status;not null" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
