package handler

import (
	"aliagha/config"
	"aliagha/services"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/eapache/go-resiliency/breaker"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type FlightTestSuite struct {
	suite.Suite
	flights   []services.FlightResponse
	f         *Flight
	redis     *redis.Client
	Validator *validator.Validate
	e         *echo.Echo
}

func (suite *FlightTestSuite) SetupTest() {

	vldt := validator.New()
	suite.e = echo.New()
	// Mock the necessary dependencies
	suite.f = &Flight{
		Redis:     &redis.Client{},
		Validator: vldt,
		Config:    &config.Config{Redis: config.Redis{TTL: 10 * time.Second}, MockAPI: config.MockAPI{Request_timeout: 30 * time.Second}},
		APIMock: services.APIMockClient{
			Client:  &http.Client{},
			Breaker: &breaker.Breaker{},
		},
	}
	// Initialize the flights for testing
	suite.flights = []services.FlightResponse{
		{
			ID: 1, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 13, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 1, Name: "Boeing 737"}, Airline: "Airline X", Price: 200, CxlSitID: 123, RemainingSeats: 50,
		},
		{
			ID: 2, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 14, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 17, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 2, Name: "Airbus A320"}, Airline: "Airline Y", Price: 250, CxlSitID: 456, RemainingSeats: 30,
		},
	}

}
func (suite *FlightTestSuite) CallHandler(requestBody string, endPoint string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, endPoint, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := suite.e.NewContext(req, res)

	var err error
	if endPoint == "/flights" {
		err = suite.f.Get(c)
	}

	if err != nil {
		return res, err
	}

	return res, nil
}
func (suite *FlightTestSuite) TestFlighGet_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK
	expectedResponse, err := json.MarshalIndent(suite.flights, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	monkey.Patch(suite.f.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.f.Validator.Struct)

	// Mock the Redis Get method to return a cache miss
	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get", func(r *redis.Client, key string) *redis.StringCmd {
		return redis.NewStringResult("", redis.Nil)
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get")

	// Mock the APIMock GetFlights method to return a sample flight response
	mockFlightResponse := suite.flights

	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.APIMock.GetFlights), "GetFlights", func(a *services.APIMockClient, departureCity, arrivalCity string, date time.Time) ([]services.FlightResponse, error) {
		return mockFlightResponse, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.APIMock.GetFlights), "GetFlights")

	requestBody := `"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28"}`

	res, err := suite.CallHandler(requestBody, "/flights")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.JSONEq(string(expectedResponse), res.Body.String())
}
func (suite *FlightTestSuite) TestFlight_BindFailure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest

	requestBody := `{"departure_city": "City A", "arrival_city": "City B"}` // Missing a requied field

	res, err := suite.CallHandler(requestBody, "/flights")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}
func (suite *FlightTestSuite) TestFlight_ValidationFailure() {
	require := suite.Require()

	tests := []struct {
		requestBody string
		statusCode  int
	}{
		{`{"departure_city": "", "arrival_city": "City B", "date": "2023-06-28"}`, http.StatusBadRequest},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28"}`, http.StatusBadRequest},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": ""}`, http.StatusBadRequest},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "some string"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		testname := tt.requestBody
		suite.T().Run(testname, func(t *testing.T) {
			res, err := suite.CallHandler(tt.requestBody, "/flights")
			require.NoError(err)
			require.Equal(tt.statusCode, res.Code)
		})
	}
}
func (suite *FlightTestSuite) TestFlight_RedisFailure() {
	require := suite.Require()
	expectedStatusCode := http.StatusInternalServerError

	// Mock the Redis Get method to return an error
	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get", func(r *redis.Client, key string) *redis.StringCmd {
		return redis.NewStringResult("", errors.New("Redis connection failed"))
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get")

	requestBody := `{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28"}`

	res, err := suite.CallHandler(requestBody, "/flights")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}
func (suite *FlightTestSuite) TestFlight_SortAndFilter() {
	require := suite.Require()
	response := make([]json.RawMessage, len(suite.flights))
	for i, flight := range suite.flights {
		jsonValue, err := json.MarshalIndent(flight, "", "  ")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		response[i] = jsonValue
	}

	monkey.Patch(suite.f.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.f.Validator.Struct)

	// Mock the Redis Get method to return a cache miss
	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get", func(r *redis.Client, key string) *redis.StringCmd {
		return redis.NewStringResult("", redis.Nil)
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.Redis), "Get")

	// Mock the APIMock GetFlights method to return a sample flight response
	mockFlightResponse := suite.flights

	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights", func(a *services.APIMockClient, departureCity, arrivalCity string, date time.Time) ([]services.FlightResponse, error) {
		return mockFlightResponse, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights")

	tests := []struct {
		requestBody string
		statusCode  int
		response    string
	}{
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "sort_by": "price", "sort_order": "asc", "remainingSeats": 20}`, http.StatusOK, string(string(response[0])) + string(response[1])},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "sort_by": "dep_time", "sort_order": "desc"}`, http.StatusOK, string(response[1]) + string(string(response[0]))},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "sort_by": "duration", "sort_order": ""}`, http.StatusOK, string(response[1]) + string(string(response[0]))},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "remainingSeats": 40}`, http.StatusOK, string(string(response[0]))},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "airline": "Airline X"}`, http.StatusOK, string(string(response[0]))},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "airplane_name": "Boeing 737"}`, http.StatusOK, string(response[0])},
		{`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "departure_time_from": "09:00:00Z"},"departure_time_to": "14:00:00Z"}`, http.StatusOK, string(response[0])},
	}

	for _, tt := range tests {
		testname := tt.requestBody
		suite.T().Run(testname, func(t *testing.T) {
			res, err := suite.CallHandler(tt.requestBody, "/flights")
			require.NoError(err)
			require.Equal(tt.statusCode, res.Code)
			require.JSONEq(tt.response, res.Body.String())
		})
	}
}

func TestFlightSuite(t *testing.T) {
	suite.Run(t, new(FlightTestSuite))
}
