package handler

import (
	"aliagha/services"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
)

type FlightReservation struct {
	DB        *gorm.DB
	Validator *validator.Validate
	APIMock   services.APIMockClient
}

type FlightReservationRequest struct {
	UserId       int   `json:"-"`
	FlightId     int   `json:"flight_id" validate:"required"`
	PassengerIds []int `json:"passenger_ids" validate:"required"`
}

func (f *FlightReservation) Reserve(ctx echo.Context) error {
	var req FlightReservationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := f.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	return ctx.JSON(http.StatusOK, req)
}
