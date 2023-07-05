package handler

import (
	"aliagha/models"
	"aliagha/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Passenger struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

type CreatePassengerRequest struct {
	Name         string `json:"name" validate:"required,min=3,max=100"`
	NationalCode string `json:"national_code" validate:"required,numeric"`
	Birthdate    string `json:"birth_date" validate:"required"`
}

type CreatePassengerResponse struct {
	Message string `json:"message"`
}

func (p *Passenger) CreatePassenger(ctx echo.Context) error {
	UID, err := strconv.Atoi(ctx.Get("user_id").(string))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	var req CreatePassengerRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := p.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	var passenger models.Passenger
	err = p.DB.Model(&models.Passenger{}).Where("u_id = ? AND national_code = ?", UID, req.NationalCode).First(&passenger).Error

	if err == nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "passenger already exists")
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(http.StatusInternalServerError, "Internal server error")
	}

	birthDate, _ := utils.ParseDate(req.Birthdate)

	passenger = models.Passenger{
		UID:          (int32)(UID),
		Name:         req.Name,
		NationalCode: req.NationalCode,
		Birthdate:    birthDate,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if p.DB.Create(&passenger).Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to create passenger")
	}

	return ctx.JSON(http.StatusCreated, CreatePassengerResponse{
		Message: "Passenger created successfully",
	})
}

type GetPassengersResponse struct {
	Passengers []PassengerResponse `json:"passengers"`
}

type PassengerResponse struct {
	ID           int32  `json:"id"`
	UID          int32  `json:"u_id"`
	NationalCode string `json:"national_code"`
	Name         string `json:"name"`
	Birthdate    string `json:"birth_date"`
}

func (p *Passenger) GetPassengers(ctx echo.Context) error {
	UID, err := strconv.Atoi(ctx.Get("user_id").(string))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	var passengers []models.Passenger
	result := p.DB.Model(&models.Passenger{}).Where("u_id = ?", UID).Find(&passengers)

	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve passengers")
	}

	resp := make([]PassengerResponse, 0, len(passengers))
	for _, passenger := range passengers {
		resp = append(resp, PassengerResponse{
			ID:           passenger.ID,
			UID:          passenger.UID,
			NationalCode: passenger.NationalCode,
			Name:         passenger.Name,
			Birthdate:    passenger.Birthdate.(string),
		})
	}
	return ctx.JSON(http.StatusOK, GetPassengersResponse{
		Passengers: passengers,
	})
}
