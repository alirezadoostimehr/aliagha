package handler

import (
	"aliagha/models"
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
	UserId       int   `json:"user_id"` // Should be taken from context
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

	//req.UserId = ctx.Get("user_id").(int)
	for _, passengerId := range req.PassengerIds {
		var exists bool
		err := f.DB.
			Debug().
			Model(&models.Passenger{}).
			Select("count(*) > 0").
			Where("id = ? AND u_id = ?", passengerId, req.UserId).
			First(&exists).
			Error
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		if exists == false {
			return ctx.JSON(http.StatusBadRequest, "Passengers Not Allowed")
		}
	}

	return ctx.JSON(http.StatusOK, req)
}
