package handler

import (
	"aliagha/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TicketTestSuite struct {
	suite.Suite
	ticket     *Ticket
	tickets    []models.Ticket
	passengers []models.Passenger
	flights    []models.Flight
	sqlMock    sqlmock.Sqlmock
	e          *echo.Echo
}

func (suite *TicketTestSuite) SetupSuite() {
	mockDB, sqlMock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))

	if err != nil {
		log.Fatal(err)
	}

	suite.sqlMock = sqlMock
	suite.e = echo.New()
	suite.ticket = &Ticket{
		DB: db,
	}

	suite.tickets = []models.Ticket{
		{ID: 100, UID: 1, PIDs: "1", FID: 235, Flight: suite.flights[0], Status: "OK", Price: 120},
		{ID: 200, UID: 1, PIDs: "2", FID: 678, Flight: suite.flights[1], Status: "Reserved", Price: 180},
		{ID: 300, UID: 1, PIDs: "1,2", FID: 910, Flight: suite.flights[2], Status: "Pending", Price: 400},
	}

	bd1, _ := time.Parse("2001-02-03", "2001-02-03")
	bd2, _ := time.Parse("2001-02-03", "2003-02-01")
	suite.passengers = []models.Passenger{
		{UID: 1, ID: 1, Name: "John Smith", NationalCode: "1234567890", Birthdate: bd1},
		{UID: 1, ID: 2, Name: "Jane Doe", NationalCode: "0123456789", Birthdate: bd2},
	}

	suite.flights = []models.Flight{
		{ID: 235, DepCityID: 1, ArrCityID: 10, AirplaneID: 25, Airline: "QatarAirways", Price: 120, CxlSitID: 2},
		{ID: 678, DepCityID: 10, ArrCityID: 1, AirplaneID: 50, Airline: "Emirates", Price: 180, CxlSitID: 1},
		{ID: 910, DepCityID: 2, ArrCityID: 3, AirplaneID: 15, Airline: "Lufthansa", Price: 200, CxlSitID: 5},
	}
}

func (suite *TicketTestSuite) CallHandler() (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, "/tickets", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := suite.e.NewContext(req, res)

	c.Set("user_id", "1")

	err := suite.ticket.GetTickets(c)
	if err != nil {
		return res, err
	}

	return res, nil
}
func (suite *TicketTestSuite) TestGetTickets_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK

	suite.sqlMock.ExpectQuery("^SELECT \\* FROM `tickets` WHERE u_id \\= \\?$").
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "u_id", "p_ids", "f_id", "status", "price"}).
			AddRow(suite.tickets[0].ID, suite.tickets[0].UID, suite.tickets[0].PIDs, suite.tickets[0].FID, suite.tickets[0].Status, suite.tickets[0].Price).
			AddRow(suite.tickets[1].ID, suite.tickets[1].UID, suite.tickets[1].PIDs, suite.tickets[1].FID, suite.tickets[1].Status, suite.tickets[1].Price).
			AddRow(suite.tickets[2].ID, suite.tickets[2].UID, suite.tickets[2].PIDs, suite.tickets[2].FID, suite.tickets[2].Status, suite.tickets[2].Price))

	suite.sqlMock.ExpectQuery("^SELECT \\* FROM `passengers` WHERE u_id \\= \\? AND id IN \\? $").
		WithArgs(0, "1,2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "national_code", "birthdate"}).
			AddRow(suite.passengers[0].ID, suite.passengers[0].Name, suite.passengers[0].NationalCode, suite.passengers[0].Birthdate).
			AddRow(suite.passengers[1].ID, suite.passengers[1].Name, suite.passengers[1].NationalCode, suite.passengers[1].Birthdate))

	suite.sqlMock.ExpectQuery(`SELECT \\* FROM "flights" WHERE id \\=\\? `).
		WithArgs(suite.flights[0].ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dep_city_id", "arr_city_id", "dep_time", "arr_time", "airplane_id", "airline", "price", "cxl_sit_id"}).
			AddRow(suite.flights[0].ID, suite.flights[0].DepCityID, suite.flights[0].ArrCityID, suite.flights[0].DepTime, suite.flights[0].ArrTime, suite.flights[0].AirplaneID, suite.flights[0].Airline, suite.flights[0].Price, suite.flights[0].CxlSitID))

	res, err := suite.CallHandler()
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)

	var response struct {
		Tickets []models.Ticket `json:"tickets"`
	}

	err = json.Unmarshal(res.Body.Bytes(), &response)

	for i := range response.Tickets {
		require.Equal(suite.tickets[i].ID, response.Tickets[i].ID)
		require.Equal(suite.tickets[i].UID, response.Tickets[i].UID)
		require.Equal(suite.tickets[i].PIDs, response.Tickets[i].PIDs)
		require.Equal(suite.tickets[i].FID, response.Tickets[i].FID)
		require.Equal(suite.tickets[i].Status, response.Tickets[i].Status)
		require.Equal(suite.tickets[i].Price, response.Tickets[i].Price)
		require.Equal(suite.tickets[i].Flight, response.Tickets[i].Flight)
	}
}

func (suite *TicketTestSuite) TestTicket_Database_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusInternalServerError
	expectedResponse := "\"Failed to retrieve tickets\""

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `tickets` WHERE u_id = (.+)").
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	res, err := suite.CallHandler()
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}
