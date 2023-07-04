package models

import "time"

type Ticket struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID       int32     `gorm:"column:u_id;not null" json:"u_id"`
	User      User      `gorm:"foreignKey:UID"`
	PIDs      string    `gorm:"column:p_ids;not null" json:"p_ids"`
	Passenger Passenger `gorm:"foreignKey:PID"`
	FID       int32     `gorm:"column:f_id;not null" json:"f_id"`
	Flight    Flight    `gorm:"foreignKey:FID"`
	Status    string    `gorm:"column:status;not null" json:"status"`
	Price     int32     `gorm:"column:price;not null" json:"price"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
