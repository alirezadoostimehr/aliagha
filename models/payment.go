package models

import "time"

type Payment struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID       int32     `gorm:"column:u_id;not null" json:"u_id"`
	User      User      `gorm:"foreignKey:UID"`
	Type      string    `gorm:"column:type;not null" json:"type"`
	TicketID  int32     `gorm:"column:ticket_id;not null" json:"ticket_id"`
	Ticket    Ticket    `gorm:"foreignKey:TicketID"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
