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
	UID       int32
}

type CreatePassengerRequest struct {
	Name         string    `json:"name" validate:"required,min=3,max=100"`
	NationalCode int32     `json:"national_code" validate:"required"`
	Birthdate    time.Time `json:"birthdate" validate:"required,date"`
}

type CreatePassengerResponse struct {
	Message string `json:"message"`
}

func (p *Passenger) CreatePassenger(ctx echo.Context) error {
	p.UID = ctx.Get("user_id").(int32)
	var req CreatePassengerRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := p.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	var passenger models.Passenger
	err := p.DB.Model(&models.Passenger{}).Where("national_code = ?", req.NationalCode).First(&passenger).Error

	if err == nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "passenger already exists")
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(http.StatusInternalServerError, "Internal server error")
	}

	passenger = models.Passenger{
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
	var passengers []models.Passenger
	result := p.DB.Model(&models.Passenger{}).Where("U_id = ?", p.UID).Find(&passengers)

	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve passengers")
	}

	return ctx.JSON(http.StatusOK, GetPassengersResponse{
		Passengers: passengers,
	})
}

type UpdatePassengerRequest struct {
	Name         string    `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	NationalCode int32     `json:"national_code,omitempty"`
	Birthdate    time.Time `json:"birthdate,omitempty"`
}

type UpdatePassengerResponse struct {
	Message string `json:"message"`
}

func (p *Passenger) UpdatePassenger(ctx echo.Context) error {
	passengerID := ctx.Param("id")

	var passenger models.Passenger
	result := p.DB.Where("u_id = ? AND id = ?", p.UID, passengerID).First(&passenger)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return ctx.JSON(http.StatusNotFound, "Passenger not found")
		} else {
			return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}

	// Update passenger fields if provided
	var req UpdatePassengerRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if req.Name != "" {
		passenger.Name = req.Name
	}
	if req.NationalCode != 0 {
		passenger.NationalCode = req.NationalCode
	}
	if !req.Birthdate.IsZero() {
		passenger.Birthdate = req.Birthdate
	}
	passenger.UpdatedAt = time.Now()

	result = p.DB.Save(&passenger)
	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to update passenger")
	}

	return ctx.JSON(http.StatusOK, UpdatePassengerResponse{
		Message: "Passenger updated successfully",
	})
}
