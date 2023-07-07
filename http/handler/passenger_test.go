package handler

import (
	"aliagha/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type PassengerTestSuite struct {
	suite.Suite
	passenger  *Passenger
	passengers []models.Passenger
	sqlMock    sqlmock.Sqlmock
	Validator  *validator.Validate
	e          *echo.Echo
}

func (suite *PassengerTestSuite) SetupSuite() {
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
	vldt := validator.New()
	suite.e = echo.New()
	suite.passenger = &Passenger{
		DB:        db,
		Validator: vldt,
	}

	bd1, _ := time.Parse("2001-02-03", "2001-02-03")
	bd2, _ := time.Parse("2001-02-03", "2003-02-01")
	suite.passengers = []models.Passenger{
		{UID: 0, ID: 1, Name: "John Smith", NationalCode: "1234567890", Birthdate: bd1},
		{UID: 0, ID: 2, Name: "Jane Doe", NationalCode: "0123456789", Birthdate: bd2},
	}

}

func (suite *PassengerTestSuite) CallHandler(requestBody string, method string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, "/passengers", strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := suite.e.NewContext(req, res)
	c.Set("user_id", "0")
	var err error

	if method == "Post" {
		err = suite.passenger.CreatePassenger(c)
	} else {
		err = suite.passenger.GetPassengers(c)
	}

	if err != nil {
		return res, err
	}

	return res, nil
}

func (suite *PassengerTestSuite) TestGetPassenger_BindErr_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest
	res, err := suite.CallHandler(`{"Jane Doe","national_code":"0123456789","birth_date":"2003-02-01"}`, "Post")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *PassengerTestSuite) TestGetPassenger_ValidationErr_Failure() {
	require := suite.Require()

	tests := []struct {
		requestBody string
		method      string
		statusCode  int
	}{
		{`{"name":"Ja","national_code":"0123456789","birth_date":"2003-02-01"}`, "Post", http.StatusUnprocessableEntity},
		{`{"name":"","national_code":"0123456789","birth_date":"2003-02-01"}`, "Post", http.StatusUnprocessableEntity},
		{`{"name":"Jane Doe","birth_date":"2003-02-01"}`, "Post", http.StatusUnprocessableEntity},
		{`{"name":"Jane Doe","national_code":"0123456789","birth_date":"2003-02"}`, "Post", http.StatusUnprocessableEntity},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.requestBody, t.method)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
	}
}
func (suite *PassengerTestSuite) TestCreatePassenger_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusCreated

	expectedResponse := `{"message": "Passenger created successfully"}`

	monkey.Patch(suite.passenger.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.passenger.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `passengers` WHERE u_id = (.+) AND national_code = (.+)  ORDER BY `passengers`.`id` LIMIT 1").
		WithArgs(0, "0123456789").
		WillReturnError(gorm.ErrRecordNotFound)

	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec("INSERT INTO `passengers`").
		WithArgs(0, "0123456789", "Jane Doe", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	res, err := suite.CallHandler(`{"name":"Jane Doe","national_code":"0123456789","birth_date":"2003-02-01"}`, "Post")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.JSONEq(expectedResponse, res.Body.String())
}
func (suite *PassengerTestSuite) TestGetPassengers_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK

	suite.sqlMock.ExpectQuery("^SELECT \\* FROM `passengers` WHERE u_id \\= \\?$").
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "national_code", "birthdate"}).
			AddRow(suite.passengers[0].ID, suite.passengers[0].Name, suite.passengers[0].NationalCode, suite.passengers[0].Birthdate).
			AddRow(suite.passengers[1].ID, suite.passengers[1].Name, suite.passengers[1].NationalCode, suite.passengers[1].Birthdate))

	res, err := suite.CallHandler("", "Get")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	var response struct {
		Passengers []models.Passenger `json:"passengers"`
	}

	err = json.Unmarshal(res.Body.Bytes(), &response)

	for i := range response.Passengers {
		require.Equal(suite.passengers[i].ID, response.Passengers[i].ID)
		require.Equal(suite.passengers[i].UID, response.Passengers[i].UID)
		require.Equal(suite.passengers[i].Name, response.Passengers[i].Name)
		require.Equal(suite.passengers[i].NationalCode, response.Passengers[i].NationalCode)
		require.Equal(suite.passengers[i].Birthdate, response.Passengers[i].Birthdate)
	}
}

func (suite *PassengerTestSuite) TestPassenger_Database_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusInternalServerError
	expectedResponse := "\"Failed to retrieve passengers\""

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `passengers` WHERE u_id = (.+)").
		WithArgs(0).
		WillReturnError(errors.New("database error"))

	res, err := suite.CallHandler("", "Get")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}
func (suite *PassengerTestSuite) TestCreatePassenger_Failure_PassengerAlreadyExists() {
	require := suite.Require()
	expectedStatusCode := http.StatusUnprocessableEntity
	expectedResponse := "\"Passenger already exists\""

	monkey.Patch(suite.passenger.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.passenger.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT \\* FROM `passengers` WHERE u_id = \\? AND national_code = \\? ORDER BY `passengers`.`id` LIMIT 1$").
		WithArgs(0, "0123456789").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "national_code", "birthdate"}).
			AddRow(suite.passengers[1].ID, suite.passengers[1].Name, suite.passengers[1].NationalCode, suite.passengers[1].Birthdate))

	res, err := suite.CallHandler(`{"name":"Jane Doe","national_code":"0123456789","birth_date":"2003-02-01"}`, "Post")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func TestPassenger(t *testing.T) {
	suite.Run(t, new(PassengerTestSuite))
}
