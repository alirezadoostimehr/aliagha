package handler

import (
	"aliagha/config"
	"aliagha/models"
	"aliagha/services"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlightTestSuite struct {
	suite.Suite
	flights   []services.FlightResponse
	Config    *config.Config
	Redis     *redis.Client
	Validator *validator.Validate
}

func TestGetFlights(t *testing.T) {
	e := echo.New()

	// Create a mock Redis client
	redisClient := redis.NewClient(&redis.Options{})
	// Create a mock validator
	validator := validator.New()

	flight := &Flight{
		Redis:     redisClient,
		Validator: validator,
	}

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/flights?departure_city=CityA&arrival_city=CityB&date=2023-06-17", nil)

	// Create a new HTTP response recorder
	rec := httptest.NewRecorder()

	// Create a new Echo context
	c := e.NewContext(req, rec)

	err := flight.Get(c)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the response status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the response body
	var flights []models.Flight
	err = json.Unmarshal(rec.Body.Bytes(), &flights)

	// Assert that the response body was successfully parsed
	assert.NoError(t, err)

	// Assert that the flights slice is not empty
	assert.NotEmpty(t, flights)
}

func TestFlight_Get(t *testing.T) {
	// Create a mock Redis and validator client
	redisClient := redis.NewClient(&redis.Options{})
	validator := validator.New()

	// Create a new instance of the FlightResponse struct

	f := &Flight{
		Config:    &config.Config{},
		Validator: validator,   // Initialize the validator with a validator instance,
		Redis:     redisClient, // Initialize the Redis client with a mock Redis client,
	}

	// Create a new Echo context for testing
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set the request body and bind it to the GetRequest struct
	reqBody := GetRequest{
		DepartureCity: services.City{Name: "CityA"},
		ArrivalCity:   services.City{Name: "CityB"},
		Date:          time.Now(),
	}
	reqJSON, _ := json.Marshal(reqBody)
	req.Body = ioutil.NopCloser(bytes.NewReader(reqJSON))
	c.SetRequest(req)

	// Call the Get function and check the response
	err := f.Get(c)
	if err != nil {
		t.Errorf("Get returned an error: %v", err)
	}

	// Assert the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Get returned a non-200 status code: %d", rec.Code)
	}

	// TODO: Assert the response body

	// more test cases
}

// func TestGetFlightsFromAPI(t *testing.T) {
// 	// Create test data
// 	depCity := services.City{Name: "CityA"}
// 	arrCity := services.City{Name: "CityB"}
// 	date := time.Now()

// 	// Create a mock HTTP server
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Assert the request URL and query parameters if needed
// 		// ...

// 		// Create a mock API response
// 		apiResult := []services.FlightResponse{
// 			// ...
// 		}

// 		// Marshal the API result to JSON and write it to the response
// 		jsonData, _ := json.Marshal(apiResult)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write(jsonData)
// 	}))
// 	defer server.Close()

// 	// Override the API URL with the mock server URL
// 	oldURL := apiURL
// 	apiURL = server.URL
// 	defer func() { apiURL = oldURL }()

// 	// Call the getFlightsFromAPI function and check the result
// 	apiResult, err := getFlightsFromAPI(depCity, arrCity, date)
// 	if err != nil {
// 		t.Errorf("getFlightsFromAPI returned an error: %v", err)
// 	}

// 	// TOOD: Assert the API result

// 	// Add more test cases as needed

// }
func TestGetRequestValidation(t *testing.T) {
	// Create a new instance of the validator
	validate := validator.New()

	// Create a valid GetRequest instance for testing
	validRequest := GetRequest{
		DepartureCity:  services.City{ID: 1, Name: "City A"},
		ArrivalCity:    services.City{ID: 2, Name: "City B"},
		Date:           time.Now(),
		Airline:        "Airline X",
		Name:           "FlightResponse XYZ",
		Deptime:        time.Now(),
		SortBy:         "price",
		SortOrder:      "asc",
		RemainingSeats: 2,
	}

	// Validate the valid request
	if err := validate.Struct(validRequest); err != nil {
		t.Errorf("Validation failed for a valid request: %v", err)
	}

	// Create an invalid GetRequest instance for testing
	invalidRequest := GetRequest{
		DepartureCity:  services.City{ID: 1, Name: ""},
		ArrivalCity:    services.City{ID: 2, Name: "City B"},
		Date:           time.Now(),
		Airline:        "Airline X",
		Name:           "FlightResponse XYZ",
		Deptime:        time.Now(),
		SortBy:         "price",
		SortOrder:      "asc",
		RemainingSeats: 2,
	}

	// Validate the invalid request
	if err := validate.Struct(invalidRequest); err == nil {
		t.Error("Validation passed for an invalid request")
	}
}

func (suite *FlightTestSuite) SetupTest() {
	// Initialize the flights for testing
	suite.flights = []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 200, DepTime: time.Date(2023, 6, 26, 9, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), Airline: "Airline B", Airplane: services.Airplane{Name: "Plane B"}, RemainingSeats: 50},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
	}
}

func (suite *FlightTestSuite) TestSortFlight_Price_Ascending() {
	expected := []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
		{Price: 200, DepTime: time.Date(2023, 6, 26, 9, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), Airline: "Airline B", Airplane: services.Airplane{Name: "Plane B"}, RemainingSeats: 50},
	}

	sortedFlights, err := sortFlight(suite.flights, "price", "asc")
	suite.NoError(err)
	suite.Equal(expected, sortedFlights)
}

func (suite *FlightTestSuite) TestSortFlight_DepTime_Descending() {
	expected := []services.FlightResponse{
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 200, DepTime: time.Date(2023, 6, 26, 9, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), Airline: "Airline B", Airplane: services.Airplane{Name: "Plane B"}, RemainingSeats: 50},
	}

	sortedFlights, err := sortFlight(suite.flights, "dep_time", "desc")
	suite.NoError(err)
	suite.Equal(expected, sortedFlights)
}

func (suite *FlightTestSuite) TestSortFlight_Duration_Ascending() {
	expected := []services.FlightResponse{
		{Price: 200, DepTime: time.Date(2023, 6, 26, 9, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), Airline: "Airline B", Airplane: services.Airplane{Name: "Plane B"}, RemainingSeats: 50},
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
	}

	sortedFlights, err := sortFlight(suite.flights, "duration", "asc")
	suite.NoError(err)
	suite.Equal(expected, sortedFlights)
}

func (suite *FlightTestSuite) TestSortFlight_InvalidSortBy() {
	_, err := sortFlight(suite.flights, "invalid_sort_by", "asc")
	suite.EqualError(err, "Invalid sort_by parameter")
}

func (suite *FlightTestSuite) TestFilterByAirline() {
	expected := []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
	}

	filteredFlights := filterByAirline(suite.flights, "Airline A")
	suite.Equal(expected, filteredFlights)
}

func (suite *FlightTestSuite) TestFilterByName() {
	expected := []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
	}

	filteredFlights := filterByName(suite.flights, "Plane A")
	suite.Equal(expected, filteredFlights)
}

func (suite *FlightTestSuite) TestFilterByDeptime() {
	expected := []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
	}

	filteredFlights := filterByDeptime(suite.flights, time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC))
	suite.Equal(expected, filteredFlights)
}

func (suite *FlightTestSuite) TestFilterByRemainingSeats() {
	expected := []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
	}

	filteredFlights := filterByRemainingSeats(suite.flights, 75)
	suite.Equal(expected, filteredFlights)
}

func TestFlightTestSuite(t *testing.T) {
	suite.Run(t, new(FlightTestSuite))
}
