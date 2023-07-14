package handler

import (
	"aliagha/config"
	"aliagha/database"
	"aliagha/services"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"bou.ke/monkey"
	"github.com/eapache/go-resiliency/breaker"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type GetFlightTestSuite struct {
	suite.Suite
	flights     []services.FlightResponse
	flight      *Flight
	redis       *redis.Client
	Validator   *validator.Validate
	e           *echo.Echo
	redisServer *miniredis.Miniredis
}

func (suite *GetFlightTestSuite) SetupSuite() {
	vldt := validator.New()
	suite.e = echo.New()
	suite.flight = &Flight{
		Redis:     &redis.Client{},
		Validator: vldt,
		Config:    &config.Config{Redis: config.Redis{TTL: 10 * time.Second}, MockAPI: config.MockAPI{Timeout: 30 * time.Second}},
		APIMock: services.APIMockClient{
			Client:  &http.Client{},
			Breaker: &breaker.Breaker{},
			Timeout: 30 * time.Second,
		},
	}
	suite.flights = []services.FlightResponse{
		{
			ID: 1, DepCity: services.City{ID: 1, Name: "CityA"}, ArrCity: services.City{ID: 2, Name: "CityB"}, DepTime: time.Date(2023, 6, 28, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 13, 0, 0, 0, time.UTC),
			Airplane: services.Airplane{ID: 1, Name: "Boeing737"}, Airline: "AirlineX", Price: 200, CxlSitID: 123, RemainingSeats: 50,
		},
		{
			ID: 2, DepCity: services.City{ID: 1, Name: "CityA"}, ArrCity: services.City{ID: 2, Name: "CityB"}, DepTime: time.Date(2023, 6, 28, 14, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 17, 0, 0, 0, time.UTC),
			Airplane: services.Airplane{ID: 2, Name: "AirbusA320"}, Airline: "AirlineY", Price: 250, CxlSitID: 456, RemainingSeats: 30,
		},
	}

	server, client := database.NewRedisMock()

	suite.redisServer = server
	suite.flight.Redis = client
}

func (suite *GetFlightTestSuite) SetupTest() {
	suite.flight.Redis.FlushAll()
}

func (suite *GetFlightTestSuite) TearDownSuite() {
	suite.redisServer.Close()
}

func (suite *GetFlightTestSuite) CallHandler(query string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodGet, "/flights"+query, bytes.NewReader([]byte("")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := suite.e.NewContext(req, res)

	err := suite.flight.Get(c)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (suite *GetFlightTestSuite) TestGetFlight_NoCache_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK
	expectedResponse := `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(_ *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	res, err := suite.CallHandler(`?departure_city=CityA&arrival_city=CityB&date=2023-06-28`)
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))

	cache, err := suite.redisServer.Get("flights-CityA-CityB-2023-06-28")
	require.NoError(err)
	require.Equal(cache, expectedResponse)
}

func (suite *GetFlightTestSuite) TestGetFlight_WithCache_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK
	expectedResponse := `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`

	monkey.Patch(suite.flight.Validator.Struct, func(_ interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.flight.Validator.Struct)

	err := suite.redisServer.Set("flights-CityA-CityB-2023-06-28", `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`)
	require.NoError(err)

	res, err := suite.CallHandler(`?departure_city=CityA&arrival_city=CityB&date=2023-06-28`)
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func (suite *GetFlightTestSuite) TestGetFlight_WithSortAndFilter_Success() {
	require := suite.Require()

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(a *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	tests := []struct {
		query      string
		statusCode int
		response   string
	}{
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&sort_by=price&sort_order=desc`, http.StatusOK, `[{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30},{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&sort_by=dep_time&sort_order=asc`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&sort_by=duration`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&remaining_seats=40`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50}]`},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
		require.Equal(t.response, strings.TrimSpace(res.Body.String()))
	}
}

func (suite *GetFlightTestSuite) TestGetFlight_BindErr_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest

	res, err := suite.CallHandler(`?departure=CityA&arrival_city=CityB&date=2023-06-28`)
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *GetFlightTestSuite) TestGetFlight_ValidationErr_Failure() {
	require := suite.Require()

	tests := []struct {
		query      string
		statusCode int
	}{
		{`?departure_city=&arrival_city=CityB&date=2023-06-28`, http.StatusBadRequest},
		{`?departure_city=CityA&arrival_city=&date=2023-06-28`, http.StatusBadRequest},
		{`?departure_city=CityA&arrival_city=CityB&date=`, http.StatusBadRequest},
		{`?departure_city=CityA&arrival_city=CityB&date=str`, http.StatusBadRequest},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
	}
}

func (suite *GetFlightTestSuite) TestGetFlight_RedisErr_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusInternalServerError

	monkey.PatchInstanceMethod(reflect.TypeOf(suite.flight.Redis), "Get", func(r *redis.Client, key string) *redis.StringCmd {
		return redis.NewStringResult("", errors.New("redis connection failed"))
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.flight.Redis), "Get")

	res, err := suite.CallHandler(`?departure_city=CityA&arrival_city=CityB&date=2023-06-28`)
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *GetFlightTestSuite) TestGetFlight_APIMockErr_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusInternalServerError

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(_ *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return nil, errors.New("error")
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	res, err := suite.CallHandler(`?departure_city=CityA&arrival_city=CityB&date=2023-06-28`)
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}
func (suite *GetFlightTestSuite) TestGetFlight_FilterByAirline_Success() {
	require := suite.Require()

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(a *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	tests := []struct {
		query      string
		statusCode int
		response   string
	}{
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&airline=AirlineY`, http.StatusOK, `[{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&airline=AirlineX`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50}]`},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
		require.Equal(t.response, strings.TrimSpace(res.Body.String()))
	}
}
func (suite *GetFlightTestSuite) TestGetFlight_FilterByAirplaneName_Success() {
	require := suite.Require()

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(a *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	tests := []struct {
		query      string
		statusCode int
		response   string
	}{
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&airplane_name=AirbusA320`, http.StatusOK, `[{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&airplane_name=Boeing737`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50}]`},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
		require.Equal(t.response, strings.TrimSpace(res.Body.String()))
	}
}
func (suite *GetFlightTestSuite) TestGetFlight_FilterByDeptime_Success() {
	require := suite.Require()

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(a *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	tests := []struct {
		query      string
		statusCode int
		response   string
	}{
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&departure_time_from=09:00&departure_time_to=15:00`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&departure_time_from=13:00&departure_time_to=15:00`, http.StatusOK, `[{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
		require.Equal(t.response, strings.TrimSpace(res.Body.String()))
	}
}
func (suite *GetFlightTestSuite) TestGetFlight_FilterByRemainingSeats_Success() {
	require := suite.Require()

	var a services.APIMockClient
	monkey.PatchInstanceMethod(reflect.TypeOf(&a), "GetFlights", func(a *services.APIMockClient, _, _, _ string) ([]services.FlightResponse, error) {
		return suite.flights, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(&a), "GetFlights")

	tests := []struct {
		query      string
		statusCode int
		response   string
	}{
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&remaining_seats=30`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50},{"id":2,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T14:00:00Z","arr_time":"2023-06-28T17:00:00Z","airplane":{"id":2,"name":"AirbusA320"},"airline":"AirlineY","price":250,"cxl_sit_id":456,"remaining_seats":30}]`},
		{`?departure_city=CityA&arrival_city=CityB&date=2023-06-28&remaining_seats=40`, http.StatusOK, `[{"id":1,"dep_city":{"id":1,"name":"CityA"},"arr_city":{"id":2,"name":"CityB"},"dep_time":"2023-06-28T10:00:00Z","arr_time":"2023-06-28T13:00:00Z","airplane":{"id":1,"name":"Boeing737"},"airline":"AirlineX","price":200,"cxl_sit_id":123,"remaining_seats":50}]`},
	}

	for _, t := range tests {
		res, err := suite.CallHandler(t.query)
		require.NoError(err)
		require.Equal(t.statusCode, res.Code)
		require.Equal(t.response, strings.TrimSpace(res.Body.String()))
	}
}
func TestFlight(t *testing.T) {
	suite.Run(t, new(GetFlightTestSuite))
}
