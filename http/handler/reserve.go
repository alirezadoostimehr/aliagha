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
	UserId       int32
	FlightId     int32   `json:"flight_id" validate:"required"`
	PassengerIds []int32 `json:"passenger_ids" validate:"required"`
}

type FlightReservationResponse struct {
	PaymentUrl string `json:"token"`
}

type ReserveVerificationRequest struct {
	Authority string `json:"token"`
}

func (f *FlightReservation) Reserve(ctx echo.Context) error {
	var req FlightReservationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Binding Error")
	}

	if err := f.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, error.Error)
	}

	userId, err := strconv.Atoi(ctx.Get("user_id").(string))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	req.UserId = int32(userId)
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
		flight := models.Flight{
			ID:               flightInfo.ID,
			DepCity:          models.City{ID: flightInfo.DepCity.ID, Name: flightInfo.DepCity.Name},
			ArrCity:          models.City{ID: flightInfo.ArrCity.ID, Name: flightInfo.ArrCity.Name},
			DepTime:          flightInfo.DepTime,
			ArrTime:          flightInfo.ArrTime,
			Airplane:         models.Airplane{ID: flightInfo.Airplane.ID, Name: flightInfo.Airplane.Name},
			Airline:          flightInfo.Airline,
			CxlSitID:         flightInfo.CxlSitID,
			FlightClass:      flightInfo.FlightClass,
			BaggageAllowance: flightInfo.BaggageAllowance,
			MealService:      flightInfo.MealService,
			Gate:             flightInfo.Gate,
		}
		if err := tx.Debug().Model(&models.Flight{}).Create(&flight).Error; err != nil && err != gorm.ErrDuplicatedKey {
			return err
		}

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
			UID:            req.UserId,
			Classification: "ticket",
			Ticket:         ticket,
			Status:         "pending",
		}

		if err := tx.Debug().Model(&models.Payment{}).Create(&payment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	zarinpal, err := gateways.NewZarinpal(f.ZarinpalConfig.MerchantId, f.ZarinpalConfig.SandBox)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	paymentUrl, authority, err := zarinpal.NewPaymentRequest(int(req.FlightId), f.ZarinpalConfig.CallbackUrl, "", "", "")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	if err := f.DB.Model(&models.Payment{}).Where("id = ?", payment.ID).Update("trans_id", authority).Error; err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "Payment failed")
	}

	return ctx.JSON(http.StatusOK, FlightReservationResponse{
		PaymentUrl: paymentUrl,
	})
}

func (f *FlightReservation) VerifyPayment(ctx echo.Context) error {
	var req ReserveVerificationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	zarinpal, err := gateways.NewZarinpal(f.ZarinpalConfig.MerchantId, f.ZarinpalConfig.SandBox)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	var payment models.Payment

	if err := f.DB.Model(&models.Payment{}).Where("trans_id = ?", req.Authority).First(&payment).Error; err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Payment not found")
	}

	var ticket models.Ticket
	if err := f.DB.Model(&payment).Association("Ticket").Find(&ticket); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	result, err := zarinpal.PaymentVerification(int(ticket.Price), req.Authority)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	if err := f.DB.Model(&models.Payment{}).Where("id = ?", payment.ID).Updates(map[string]interface{}{
		"ref_id": result.RefID,
		"status": "verified",
	}).Error; err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "Payment failed")
	}
	return nil
}
