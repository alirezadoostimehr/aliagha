package handler

import (
	"aliagha/models"
	"aliagha/services"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Cancel struct {
	DB        *gorm.DB
	Validator *validator.Validate
	APIMock   services.APIMockClient
}

type CancelRequest struct {
	TicketID    int32 `query:"ticket_id" validate:"required"`
	UserID      int32 `query:"user_id" validate:"required"`
	PassengerID int32 `query:"passenger_id" validate:"required"`
	FlightID    int32 `query:"flight_id" validate:"required"`
}

func (c *Cancel) Get(ctx echo.Context) error {
	var req CancelRequest
	// UID, err1 := strconv.Atoi(ctx.Get("user_id").(string))
	// if err1 != nil {
	// 	return ctx.JSON(http.StatusBadRequest, err.Error())
	// }
	// req.UserID = UID
	if err := ctx.Bind(&req); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := c.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Check if the ticket exists
	var ticket models.Ticket
	err := c.DB.Debug().Model(&models.Ticket{}).Where("u_id = ? AND p_id = ? AND id = ?", req.UserID, req.PassengerID, req.TicketID).First(&ticket).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.JSON(http.StatusNotFound, "Ticket not found")
		}
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// return_policy := ticket.Flight.CxlSitId.Data
	// price := ticket.Flight.Price
	// return_amount := return_policy * price
	// todo: add price to peyment/ticket for returning

	// payment := models.Payment{
	// 	UID:    req.UserID,
	// 	Type:   "ticket-cancelation",
	// 	Ticket: ticket,
	// 	Status: "pending",
	// }

	if err := c.APIMock.Cancel(req.FlightID); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// Update the ticket status to "cancelled"
	err = c.DB.Debug().Model(&models.Ticket{}).Where("u_id = ? AND p_id = ? AND id = ?", req.UserID, req.PassengerID, req.TicketID).Update("status", "cancelled").Error
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// Update the payment status to "cancelled"
	err = c.DB.Debug().Model(&models.Payment{}).Where("ticket_id = ? AND u_id = ?", req.TicketID, req.UserID).Update("status", "cancelled").Error
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusOK, "Ticket cancelled successfully")
}
