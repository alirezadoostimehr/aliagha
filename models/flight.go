package models

import "time"

type Flight struct {
	ID         int32              `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	DepCityID  int32              `gorm:"column:dep_city_id;not null" json:"dep_city_id"`
	DepCity    City               `gorm:"foreignKey:DepCityID"`
	ArrCityID  int32              `gorm:"column:arr_city_id;not null" json:"arr_city_id"`
	ArrCity    City               `gorm:"foreignKey:ArrCityID"`
	DepTime    time.Time          `gorm:"column:dep_time;not null" json:"dep_time"`
	ArrTime    time.Time          `gorm:"column:arr_time;not null" json:"arr_time"`
	AirplaneID int32              `gorm:"column:airplane_id;not null" json:"airplane_id"`
	Airplane   Airplane           `gorm:"foreignKey:AirplaneID"`
	Airline    string             `gorm:"column:airline;not null" json:"airline"`
	Price      int32              `gorm:"column:price;not null" json:"price"`
	CxlSitID   int32              `gorm:"column:cxl_sit_id;not null" json:"cxl_sit_id"`
	CxlSitId   CancelingSituation `gorm:"foreignKey:CxlSitID"`
	CreatedAt  time.Time          `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time          `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
