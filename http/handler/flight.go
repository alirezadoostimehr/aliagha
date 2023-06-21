package handler

import (
	"aliagha/models"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type Flight struct {
	Redis     *redis.Client
	Validator *validator.Validate
}

type GetFlightRequest struct {
	ID            int
	DepartureCity string `json:"departure_city" validate:"required"`
	ArrivalCity   string `json:"arrival_city" validate:"required"`
	Date          string `json:"date" validate:"required,date"`
	Airline       string
	Name          string
	DepTimeStr    string /*`validate:"time"`*/
	SortBy        string
	SortOrder     string
}

func (f *Flight) Get(c echo.Context) error {
	TTL := 1 * time.Minute
	var freq GetFlightRequest
	if err := c.Bind(&freq); err != nil {
		return c.JSON(http.StatusBadRequest, "")
	}
	if err := f.Validator.Struct(&freq); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", freq.DepartureCity, freq.ArrivalCity, freq.Date)
	cacheResult, err := f.Redis.Get(cacheKey).Result()
	var apiResult []models.Flight
	if err != nil && err != redis.Nil {
		// Error reading from cache
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get result from cache"})
	} else if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err = f.getFlightsFromAPI(freq.DepartureCity, freq.ArrivalCity, freq.Date)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flights from API"})
		}
		// Store result in cache
		jsonData, err := json.Marshal(apiResult)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
		}
		err = f.Redis.Set(cacheKey, jsonData, TTL).Err()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store result in cache"})
		}
	}

	// Filter
	var filteredFlights []models.Flight
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid airline ID"})
	}
	if freq.Airline != "" {
		filteredFlights = append(filteredFlights, filterAirline(apiResult, freq.Airline)...)
	}
	if freq.Name != "" {
		filteredFlights = append(filteredFlights, filterName(apiResult, freq.Name)...)
	}
	if freq.DepTimeStr != "" {
		depTime, err := time.Parse("15:04", freq.DepTimeStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid departure time format"})
		}
		filteredFlights = append(filteredFlights, filterDeptime(apiResult, depTime)...)
	}

	// Sort
	if freq.SortBy != "" {
		flights, err := sortFlight(apiResult, freq.SortBy, freq.SortOrder)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sort argument"})
		}
		filteredFlights = append(filteredFlights, flights...)
	}

	// Return filtered results
	jsonData, err := json.Marshal(filteredFlights)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
	}
	cacheResult = string(jsonData)

	var flights []models.Flight
	err = json.Unmarshal([]byte(cacheResult), &flights)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}
	return c.JSON(http.StatusOK, flights)
}
func (f *Flight) getFlightsFromAPI(depCity, arrCity, date string) ([]models.Flight, error) {

	url := fmt.Sprintf("https://github.com/kianakholousi/Flight-Data-API?DepartureCity=%s&ArrivalCity=%s&Date=%s", depCity, arrCity, date)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResult []models.Flight
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return nil, err
	}

	return apiResult, nil
}
func (f *Flight) getFlightFromAPI(c echo.Context) ([]models.Flight, error) {
	id := c.QueryParam("id")
	url := fmt.Sprintf("https://github.com/kianakholousi/Flight-Data-API/%s", id)
	var FlightModule []models.Flight
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return FlightModule, err
	}
	if err := f.Validator.Struct(&req); err != nil {
		return nil, c.JSON(http.StatusBadRequest, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return FlightModule, err
	}
	defer resp.Body.Close()

	var apiResult []models.Flight
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return FlightModule, err
	}

	return apiResult, nil
}
func sortFlight(flights []models.Flight, sortBy, sortOrder string) ([]models.Flight, error) {
	if sortBy == "" {
		sortOrder = "asc"
	}
	switch sortBy {
	case "price":
		if sortOrder == "asc" {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].Price < flights[j].Price
			})
		} else {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].Price > flights[j].Price
			})
		}
	case "dep_time":
		if sortOrder == "asc" {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].DepTime.Before(flights[j].DepTime)
			})
		} else {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].DepTime.After(flights[j].DepTime)
			})
		}
	case "duration":
		if sortOrder == "asc" {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].ArrTime.Sub(flights[i].DepTime) < flights[j].ArrTime.Sub(flights[j].DepTime)
			})
		} else {
			sort.Slice(flights, func(i, j int) bool {
				return flights[i].ArrTime.Sub(flights[i].DepTime) > flights[j].ArrTime.Sub(flights[j].DepTime)
			})
		}
	default:
		return nil, fmt.Errorf("Invalid sort_by parameter")
	}
	return flights, nil
}
func filterAirline(flights []models.Flight, filter string) []models.Flight {
	var filteredFlights []models.Flight
	for _, flight := range flights {
		if flight.Airline == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}
	return filteredFlights
}
func filterName(flights []models.Flight, filter string) []models.Flight {
	var filteredFlights []models.Flight
	for _, flight := range flights {
		if flight.Airplane.Name == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}
	return filteredFlights
}
func filterDeptime(flights []models.Flight, filter time.Time) []models.Flight {
	var filteredFlights []models.Flight
	for _, flight := range flights {
		if flight.DepTime.Hour() == filter.Hour() || flight.DepTime.Minute() == filter.Minute() {
			filteredFlights = append(filteredFlights, flight)
		}
	}
	return filteredFlights
}
