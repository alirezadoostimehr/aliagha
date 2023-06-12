package models

import "time"

type Flight struct {
	ID         int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	DepCityID  int32     `gorm:"column:dep_city_id;not null" json:"dep_city_id"`
	ArrCityID  int32     `gorm:"column:arr_city_id;not null" json:"arr_city_id"`
	DepTime    time.Time `gorm:"column:dep_time;not null" json:"dep_time"`
	ArrTime    time.Time `gorm:"column:arr_time;not null" json:"arr_time"`
	AirplaneID int32     `gorm:"column:airplane_id;not null" json:"airplane_id"`
	Price      int32     `gorm:"column:price;not null" json:"price"`
	CxlSitID   int32     `gorm:"column:cxl_sit_id;not null" json:"cxl_sit_id"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
