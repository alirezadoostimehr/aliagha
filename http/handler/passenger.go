package handler

import (
	"aliagha/config"
	"aliagha/models"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Passenger struct {
	DB        *gorm.DB
	JWT       *config.JWT
	Validator *validator.Validate
	UID       int32
}

type CreatePassengerRequest struct {
	Name         string    `json:"name" validate:"required,min=2,max=100"`
	NationalCode int32     `json:"national_code" validate:"required"`
	Birthdate    time.Time `json:"birthdate" validate:"required"`
	UID          int32     `json:"uid" validate:"required"`
}

type CreatePassengerResponse struct {
	Message string `json:"message"`
}

func (p *Passenger) SetUserID(userID int32) {
	p.UID = userID
}
func (p *Passenger) CreatePassenger(c echo.Context) error {
	var req CreatePassengerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := p.Validator.Struct(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	passenger := models.Passenger{
		Name:         req.Name,
		NationalCode: req.NationalCode,
		Birthdate:    req.Birthdate,
		UID:          req.UID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if p.DB.Create(&passenger).Error != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to create passenger")
	}

	return c.JSON(http.StatusOK, CreatePassengerResponse{
		Message: "Passenger created successfully",
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

func (p *Passenger) UpdatePassenger(c echo.Context) error {
	passengerID := c.Param("id")

	var passenger models.Passenger
	result := p.DB.Where("uid = ? AND id = ?", p.UID, passengerID).First(&passenger)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, "Passenger not found")
		} else {
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}

	// Update passenger fields if provided
	var req UpdatePassengerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
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
		return c.JSON(http.StatusInternalServerError, "Failed to update passenger")
	}

	return c.JSON(http.StatusOK, UpdatePassengerResponse{
		Message: "Passenger updated successfully",
	})
}

type ViewPassengersResponse struct {
	Passengers []models.Passenger `json:"passengers"`
	UID        int32              `json:"uid"`
}

func (p *Passenger) ViewPassengers(c echo.Context) error {
	var passengers []models.Passenger
	result := p.DB.Find(&passengers)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve passengers")
	}

	return c.JSON(http.StatusOK, ViewPassengersResponse{
		Passengers: passengers,
	})
}
