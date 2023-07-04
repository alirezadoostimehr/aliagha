package models

import "time"

type Passenger struct {
	ID           int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID          int32     `gorm:"column:u_id;not null" json:"u_id"`
	User         User      `gorm:"foreignKey:UID"`
	NationalCode string    `gorm:"column:national_code;not null" json:"national_code"`
	Name         string    `gorm:"column:name;not null" json:"name"`
	Birthdate    time.Time `gorm:"column:birthdate;not null" json:"birthdate"`
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
