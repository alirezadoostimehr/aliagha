package handler

import (
	"aliagha/models"
	"net/http"
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
	Name         string    `json:"name" validate:"required,min=3,max=100"`
	NationalCode string    `json:"national_code" validate:"required,numerical,length=10"`
	Birthdate    time.Time `json:"birthdate" validate:"required,date"`
}

type CreatePassengerResponse struct {
	Message string `json:"message"`
}

func (p *Passenger) CreatePassenger(ctx echo.Context) error {
	UID := ctx.Get("user_id").(int32)
	var req CreatePassengerRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := p.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	var passenger models.Passenger
	err := p.DB.Model(&models.Passenger{}).Where("u_id = ? AND national_code = ?", UID, req.NationalCode).First(&passenger).Error

	if err == nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "passenger already exists")
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(http.StatusInternalServerError, "Internal server error")
	}

	passenger = models.Passenger{
		UID:          UID,
		Name:         req.Name,
		NationalCode: req.NationalCode,
		Birthdate:    req.Birthdate,
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
	Passengers []models.Passenger `json:"passengers"`
}

func (p *Passenger) GetPassengers(ctx echo.Context) error {
	UID := ctx.Get("user_id").(int32)
	var passengers []models.Passenger
	result := p.DB.Model(&models.Passenger{}).Where("u_id = ?", UID).Find(&passengers)

	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve passengers")
	}

	return ctx.JSON(http.StatusOK, GetPassengersResponse{
		Passengers: passengers,
	})
}
