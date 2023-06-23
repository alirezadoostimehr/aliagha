package models

import "time"

type User struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Password  string    `gorm:"column:password;not null" json:"password"`
	Cellphone string    `gorm:"column:cellphone;not null" json:"cellphone"`
	Email     string    `gorm:"column:email;not null;uniqueIndex" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
