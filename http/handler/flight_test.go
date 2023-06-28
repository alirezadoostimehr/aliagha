package handler

import (
	"aliagha/config"
	"aliagha/services"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
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
		Config:    &config.Config{},
		APIMock:   services.APIMockClient{},
	}
	// Initialize the flights for testing
	suite.flights = []services.FlightResponse{
		{Price: 100, DepTime: time.Date(2023, 6, 26, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 12, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 100},
		{Price: 200, DepTime: time.Date(2023, 6, 26, 9, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), Airline: "Airline B", Airplane: services.Airplane{Name: "Plane B"}, RemainingSeats: 50},
		{Price: 150, DepTime: time.Date(2023, 6, 26, 11, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 26, 13, 0, 0, 0, time.UTC), Airline: "Airline A", Airplane: services.Airplane{Name: "Plane A"}, RemainingSeats: 75},
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
	expectedResponse := `[
		{
		  "ID": 1,
		  "DepCity": {
			"ID": 1,
			"AirplaneName": "City A"
		  },
		  "ArrCity": {
			"ID": 2,
			"AirplaneName": "City B"
		  },
		  "DepTime": "2023-06-28T10:00:00Z",
		  "ArrTime": "2023-06-28T13:00:00Z",
		  "Date": "2023-06-28",
		  "Airplane": {
			"ID": 1,
			"AirplaneName": "Boeing 737"
		  },
		  "Airline": "Airline X",
		  "Price": 200,
		  "CxlSitID": 123,
		  "RemainingSeats": 50
		},
		{
		  "ID": 2,
		  "DepCity": {
			"ID": 1,
			"AirplaneName": "City A"
		  },
		  "ArrCity": {
			"ID": 2,
			"AirplaneName": "City B"
		  },
		  "DepTime": "2023-06-28T14:00:00Z",
		  "ArrTime": "2023-06-28T17:00:00Z",
		  "Date": "2023-06-28",
		  "Airplane": {
			"ID": 2,
			"AirplaneName": "Airbus A320"
		  },
		  "Airline": "Airline Y",
		  "Price": 250,
		  "CxlSitID": 456,
		  "RemainingSeats": 30
		}
	  ]`

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
	mockFlightResponse := []services.FlightResponse{
		{
			ID: 1, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 13, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 1, Name: "Boeing 737"}, Airline: "Airline X", Price: 200, CxlSitID: 123, RemainingSeats: 50,
		},
		{
			ID: 2, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 14, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 17, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 2, Name: "Airbus A320"}, Airline: "Airline Y", Price: 250, CxlSitID: 456, RemainingSeats: 30,
		},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights", func(a *services.APIMockClient, departureCity, arrivalCity string, date time.Time) ([]services.FlightResponse, error) {
		return mockFlightResponse, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights")

	requestBody := `"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28"}`

	res, err := suite.CallHandler(requestBody, "/flights")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}
func (suite *FlightTestSuite) TestFlight_ValidationFailure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest
	// {"City A","City B","2022-06-28","Airline X","Airbus XYZ","2023-06-28 16:12:14","price","asc",RemainingSeats: 2}
	{
		res, err := suite.CallHandler(`{"departure_city": "", "arrival_city": "City B", "date": "2023-06-28"}`, "/flights")
		require.NoError(err)
		require.Equal(expectedStatusCode, res.Code)
	}

	{
		res, err := suite.CallHandler(`{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28"}`, "/flights")
		require.NoError(err)
		require.Equal(expectedStatusCode, res.Code)
	}

	{
		res, err := suite.CallHandler(`{"departure_city": "City A", "arrival_city": "City B", "date": ""}`, "/flights")
		require.NoError(err)
		require.Equal(expectedStatusCode, res.Code)
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
	expectedStatusCode := http.StatusOK
	expectedResponse := `[
	{
	  "ID": 1,
	  "DepCity": {
	"ID": 1,
	"Name": "City A"
	  },
	  "ArrCity": {
	"ID": 2,
	"Name": "City B"
	  },
	  "DepTime": "2023-06-28T10:00:00Z",
	  "ArrTime": "2023-06-28T13:00:00Z",
	  "Date": "2023-06-28",
	  "Airplane": {
	"ID": 1,
	"Name": "Boeing 737"
	  },
	  "Airline": "Airline X",
	  "Price": 200,
	  "CxlSitID": 123,
	  "RemainingSeats": 50
	},
	{
	  "ID": 2,
	  "DepCity": {
	"ID": 1,
	"Name": "City A"
	  },
	  "ArrCity": {
	"ID": 2,
	"Name": "City B"
	  },
	  "DepTime": "2023-06-28T14:00:00Z",
	  "ArrTime": "2023-06-28T17:00:00Z",
	  "Date": "2023-06-28",
	  "Airplane": {
	"ID": 2,
	"Name": "Airbus A320"
	  },
	  "Airline": "Airline Y",
	  "Price": 250,
	  "CxlSitID": 456,
	  "RemainingSeats": 30
	}
	  ]`

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
	mockFlightResponse := []services.FlightResponse{
		{
			ID: 1, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 10, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 13, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 1, Name: "Boeing 737"}, Airline: "Airline X", Price: 200, CxlSitID: 123, RemainingSeats: 50,
		},
		{
			ID: 2, DepCity: services.City{ID: 1, Name: "City A"}, ArrCity: services.City{ID: 2, Name: "City B"}, DepTime: time.Date(2023, 6, 28, 14, 0, 0, 0, time.UTC), ArrTime: time.Date(2023, 6, 28, 17, 0, 0, 0, time.UTC),
			Date: time.Date(2023, 6, 28, 0, 0, 0, 0, time.UTC), Airplane: services.Airplane{ID: 2, Name: "Airbus A320"}, Airline: "Airline Y", Price: 250, CxlSitID: 456, RemainingSeats: 30,
		},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights", func(a *services.APIMockClient, departureCity, arrivalCity string, date time.Time) ([]services.FlightResponse, error) {
		return mockFlightResponse, nil
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(suite.f.APIMock), "GetFlights")

	requestBody := `{"departure_city": "City A", "arrival_city": "City B", "date": "2023-06-28", "sort_by": "price", "sort_order": "asc", "remainingSeats": 2}`

	res, err := suite.CallHandler(requestBody, "/flights")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func TestFlightTestSuite(t *testing.T) {
	suite.Run(t, new(FlightTestSuite))
}
