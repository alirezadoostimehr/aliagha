package models

import "time"

type Payment struct {
	ID             int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID            int32     `gorm:"column:u_id;not null" json:"u_id"`
	TransId        *string   `gorm:"column:trans_id;null" json:"trans_id"`
	RefId          *string   `gorm:"column:ref_id;null" json:"ref_id"`
	User           User      `gorm:"foreignKey:UID"`
	Classification string    `gorm:"column:classification;not null" json:"classification"`
	TicketID       int32     `gorm:"column:ticket_id;not null" json:"ticket_id"`
	Ticket         Ticket    `gorm:"foreignKey:TicketID"`
	Status         string    `gorm:"column:status;not null" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
