package handler

import (
	"aliagha/models"
	"net/http"
	"strconv"
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

type City struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type Airplane struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type FlightResponse struct {
	ID       int32     `json:"id"`
	DepCity  City      `json:"dep_city"`
	ArrCity  City      `json:"arr_city"`
	DepTime  time.Time `json:"dep_time"`
	ArrTime  time.Time `json:"arr_time"`
	Airplane Airplane  `json:"airplane"`
	Airline  string    `json:"airline"`
	Price    int32     `json:"price"`
	CxlSitID int32     `json:"cxl_sit_id"`
}

type TicketResponse struct {
	ID        int32 `json:"id"`
	PID       int32 `json:"p_id"`
	Passenger PassengerResponse
	FID       int32 `json:"f_id"`
	Flight    FlightResponse
	Status    string `json:"status"`
}

func (t *Ticket) GetTickets(ctx echo.Context) error {
	UID, err := strconv.Atoi(ctx.Get("user_id").(string))

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	var tickets []models.Ticket
	result := t.DB.Model(&models.Ticket{}).Where("u_id = ?", UID).Find(&tickets)

	if result.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve tickets")
	}

	resp := make([]TicketResponse, 0, len(tickets))
	var passenger PassengerResponse
	var flight FlightResponse

	for _, ticket := range tickets {
		err := t.DB.Model(&models.Ticket{}).Where("p_id = ?", ticket.PID).Find(&passenger).Error
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve tickets")
		}

		err = t.DB.Model(&models.Ticket{}).Where("f_id = ?", ticket.FID).Find(&flight).Error
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Failed to retrieve flights")
		}

		resp = append(resp, TicketResponse{
			ID:        ticket.ID,
			Passenger: passenger,
			Flight:    flight,
			Status:    ticket.Status,
		})
	}
	return ctx.JSON(http.StatusOK, GetTicketsResponse{
		Tickets: resp,
	})
}
