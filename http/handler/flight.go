package handler

import (
	"aliagha/config"
	"aliagha/services"
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
	Config    *config.Config
}

type GetRequest struct {
	DepartureCity  services.City `query:"departure_city" validate:"required"`
	ArrivalCity    services.City `query:"arrival_city" validate:"required"`
	Date           time.Time     `query:"date" validate:"required,datetime"`
	Airline        string        `query:"airline"`
	Name           string        `query:"name"`
	Deptime        time.Time     `query:"departure_time"` /*validate:"datetime"*/
	SortBy         string        `query:"sort_by"`
	SortOrder      string        `query:"sort_order"`
	RemainingSeats int32         `query:"remaining_seats"`
}

func (f *Flight) Get(ctx echo.Context) error {
	TTL := f.Config.Redis.TTL

	var req GetRequest
	var flights []services.FlightModuleAPI
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	if err := f.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// req.Date = req.Deptime.Format("2006-01-02")
	cacheKey := fmt.Sprintf("%s-%s-%s", req.DepartureCity.Name, req.ArrivalCity.Name, req.Date.Format("2003-02-01"))
	cacheResult, err := f.Redis.Get(cacheKey).Bytes()

	if err != nil && err != redis.Nil {
		return ctx.JSON(http.StatusInternalServerError, "Inetrnal Server Error")
	} else if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := services.GetFlightsFromAPI(req.DepartureCity, req.ArrivalCity, req.Date)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flights from API"})
		}
		// Store result in cache
		jsonData, err := json.Marshal(apiResult)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
		}
		err = f.Redis.Set(cacheKey, jsonData, TTL).Err()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store result in cache"})
		}

		flights = apiResult
	} else {
		err = json.Unmarshal(cacheResult, &flights)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
	}

	// Filter

	if err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid airline ID"})
	}

	if req.Airline != "" {
		flights = append(flights, filterByAirline(flights, req.Airline)...)
	}

	if req.Name != "" {
		flights = append(flights, filterByName(flights, req.Name)...)
	}

	if req.Deptime.Format("2003-02-01") != "" {

		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid departure time format"})
		}

		flights = append(flights, filterByDeptime(flights, req.Deptime)...)
	}
	if req.RemainingSeats > 0 {
		flights = append(flights, filterByRemainingSeats(flights, req.RemainingSeats)...)
	}

	// Sort
	if req.SortBy != "" {
		flights, err := sortFlight(flights, req.SortBy, req.SortOrder)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sort argument"})
		}

		flights = append(flights, flights...)
	}

	// Return filtered results
	jsonData, err := json.Marshal(flights)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
	}
	cacheResult = jsonData

	err = json.Unmarshal([]byte(cacheResult), &flights)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}

	return ctx.JSON(http.StatusOK, flights)
}

func sortFlight(flights []services.FlightModuleAPI, sortBy, sortOrder string) ([]services.FlightModuleAPI, error) {
	if sortOrder == "" {
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
func filterByAirline(flights []services.FlightModuleAPI, filter string) []services.FlightModuleAPI {
	var filteredFlights []services.FlightModuleAPI
	for _, flight := range flights {
		if flight.Airline == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByName(flights []services.FlightModuleAPI, filter string) []services.FlightModuleAPI {
	var filteredFlights []services.FlightModuleAPI
	for _, flight := range flights {
		if flight.Airplane.Name == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByDeptime(flights []services.FlightModuleAPI, filter time.Time) []services.FlightModuleAPI {
	var filteredFlights []services.FlightModuleAPI
	for _, flight := range flights {
		if flight.DepTime.Hour() == filter.Hour() && flight.DepTime.Minute() == filter.Minute() {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByRemainingSeats(flights []services.FlightModuleAPI, filter int32) []services.FlightModuleAPI {
	var filteredFlights []services.FlightModuleAPI
	for _, flight := range flights {
		if flight.RemainingSeats >= filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
