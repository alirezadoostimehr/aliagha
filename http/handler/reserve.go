package handler

import (
	"aliagha/config"
	"aliagha/models"
	"aliagha/services"
	"aliagha/utils/gateways"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type FlightReservation struct {
	DB             *gorm.DB
	ZarinpalConfig *config.Zarinpal
	Validator      *validator.Validate
	APIMock        services.APIMockClient
}

type FlightReservationRequest struct {
	UserId       int32   `json:"user_id"` // Should be taken from context
	FlightId     int32   `json:"flight_id" validate:"required"`
	PassengerIds []int32 `json:"passenger_ids" validate:"required"`
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

	if err := f.APIMock.Reserve(req.FlightId, (int32)(len(req.PassengerIds))); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	flightInfo, err := f.APIMock.GetFlightInfo(req.FlightId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	passengerIdsStr := ""
	for i, passengerId := range req.PassengerIds {
		if i > 0 {
			passengerIdsStr += ", "
		}
		passengerIdsStr += strconv.Itoa((int)(passengerId))
	}

	var payment models.Payment
	err = f.DB.Debug().Transaction(func(tx *gorm.DB) error {
		ticket := models.Ticket{
			UID:    req.UserId,
			PIDs:   passengerIdsStr,
			FID:    req.FlightId,
			Status: "payment pending",
			Price:  flightInfo.Price,
		}

		if err := tx.Debug().Model(&models.Ticket{}).Create(&ticket).Error; err != nil {
			return err
		}

		payment = models.Payment{
			UID:    req.UserId,
			Type:   "ticket",
			Ticket: ticket,
			Status: "pending",
		}

		if err := tx.Debug().Model(&models.Payment{}).Create(&payment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	zarinpal, err := gateways.NewZarinpal("test", true)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	_, authority, err := zarinpal.NewPaymentRequest(int(req.FlightId), f.ZarinpalConfig.CallbackUrl, "", "", "")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	if err := f.DB.Model(&models.Payment{}).Where("id = ?", payment.ID).Update("authority", authority).Error; err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "Payment failed")
	}

	return ctx.JSON(http.StatusOK, req)
}
