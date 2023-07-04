package handler

import (
	"aliagha/services"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type FlightReservation struct {
	DB        *gorm.DB
	Validator *validator.Validate
	APIMock   services.APIMockClient
}

func (f *FlightReservation) Reserve(ctx echo.Context) error {
	return nil
}
