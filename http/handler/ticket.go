package handler

import (
	"aliagha/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Ticket struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

type GetTicketsResponse struct {
	Tickets []TicketResponse `json:"tickets"`
}

type CityResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type AirplaneResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type FlightResponse struct {
	ID       int32            `json:"id"`
	DepCity  CityResponse     `json:"dep_city"`
	ArrCity  CityResponse     `json:"arr_city"`
	DepTime  time.Time        `json:"dep_time"`
	ArrTime  time.Time        `json:"arr_time"`
	Airplane AirplaneResponse `json:"airplane"`
	Airline  string           `json:"airline"`
	Price    int32            `json:"price"`
	CxlSitID int32            `json:"cxl_sit_id"`
}

type TicketResponse struct {
	ID        int32 `json:"id"`
	Passenger []PassengerResponse
	Flight    FlightResponse `json:"flight"`
	Status    string         `json:"status"`
}

func (t *Ticket) GetTickets(ctx echo.Context) error {
	UID, err := strconv.Atoi(ctx.Get("user_id").(string))

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	var tickets []models.Ticket
	result := t.DB.Debug().Model(&models.Ticket{}).
		Where("u_id = ?", UID).
		Preload("Flight").
		Preload("User").
		Find(&tickets)

	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve tickets")
	}

	resp := make([]TicketResponse, 0, len(tickets))
	var passengers []models.Passenger
	var flight models.Flight

	for _, ticket := range tickets {

		err := t.DB.Debug().Model(&models.Passenger{}).Select("*").
			Where("u_id = ? AND id IN (?)", UID, strings.Split(ticket.PIDs, ",")).
			Find(&passengers).Error
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve passegers")
		}

		passengerResponse := make([]PassengerResponse, 0, len(passengers))
		for _, passenger := range passengers {
			passengerResponse = append(passengerResponse, PassengerResponse{
				ID:           passenger.ID,
				UID:          passenger.UID,
				Name:         passenger.Name,
				NationalCode: passenger.NationalCode,
				Birthdate:    passenger.Birthdate.Format("2003-02-01"),
			})
		}

		err = t.DB.Debug().Model(&models.Flight{}).
			Joins("Airplane").
			Joins("DepCity").Where("DepCity.id = ? ", ticket.Flight.DepCityID).
			Joins("ArrCity").Where("ArrCity.id = ? ", ticket.Flight.ArrCityID).
			Where("flights.id = ?", ticket.FID).
			First(&flight).Error
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve passegers")
		}

		flightResponse := FlightResponse{
			ID:       flight.ID,
			DepCity:  CityResponse{ID: flight.DepCity.ID, Name: flight.DepCity.Name},
			ArrCity:  CityResponse{ID: flight.ArrCity.ID, Name: flight.ArrCity.Name},
			DepTime:  flight.DepTime,
			ArrTime:  flight.ArrTime,
			Airplane: AirplaneResponse{ID: flight.Airplane.ID, Name: flight.Airplane.Name},
			Airline:  flight.Airline,
			Price:    flight.Price,
			CxlSitID: flight.CxlSitID,
		}

		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve flights")
		}

		resp = append(resp, TicketResponse{
			ID:        ticket.ID,
			Passenger: passengerResponse,
			Flight:    flightResponse,
			Status:    ticket.Status,
		})
	}
	return ctx.JSON(http.StatusOK, GetTicketsResponse{
		Tickets: resp,
	})
}
