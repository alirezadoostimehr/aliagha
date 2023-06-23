package handler

import (
	"aliagha/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

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
