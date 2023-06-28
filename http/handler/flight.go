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
	APIMock   services.APIMockClient
}

type GetRequest struct {
	DepartureCity  string    `query:"departure_city" validate:"required"`
	ArrivalCity    string    `query:"arrival_city" validate:"required"`
	Date           time.Time `query:"date" validate:"required,datetime"`
	Airline        string    `query:"airline"`
	AirplaneName   string    `query:"airplane_name"`
	DeptimeFrom    time.Time `query:"departure_time_from" validate:"datetime"`
	DeptimeTo      time.Time `query:"departure_time_to" validate:"datetime"`
	SortBy         string    `query:"sort_by"`
	SortOrder      string    `query:"sort_order"`
	RemainingSeats int32     `query:"remaining_seats"`
}

func (f *Flight) Get(ctx echo.Context) error {
	var req GetRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	if err := f.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", req.DepartureCity, req.ArrivalCity, req.Date.Format("2003-02-01"))
	cacheResult, err := f.Redis.Get(cacheKey).Bytes()

	var flights []services.FlightResponse
	if err != nil && err != redis.Nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	} else if err == redis.Nil {
		apiResult, err := f.APIMock.GetFlights(req.DepartureCity, req.ArrivalCity, req.Date)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Internal Sever Error")
		}

		jsonData, err := json.Marshal(apiResult)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}

		err = f.Redis.Set(cacheKey, jsonData, f.Config.Redis.TTL).Err()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Internal Sever Error")
		}

		flights = apiResult
	} else {
		err = json.Unmarshal(cacheResult, &flights)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
	}

	if req.Airline != "" {
		flights = filterByAirline(flights, req.Airline)
	}

	if req.AirplaneName != "" {
		flights = filterByName(flights, req.AirplaneName)
	}

	if req.DeptimeFrom.Format("2003-02-01") != "" && req.DeptimeTo.Format("2003-02-01") != "" {
		flights = filterByDeptime(flights, req.DeptimeFrom, req.DeptimeTo)
	}

	if req.RemainingSeats > 0 {
		flights = filterByRemainingSeats(flights, req.RemainingSeats)
	}

	if req.SortBy != "" {
		flights = sortFlight(flights, req.SortBy, req.SortOrder)
	}

	return ctx.JSON(http.StatusOK, flights)
}

func sortFlight(flights []services.FlightResponse, sortBy, sortOrder string) []services.FlightResponse {
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
		return flights
	}

	return flights
}

func filterByAirline(flights []services.FlightResponse, airline string) []services.FlightResponse {
	var filteredFlights []services.FlightResponse
	for _, flight := range flights {
		if flight.Airline == airline {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}

func filterByName(flights []services.FlightResponse, name string) []services.FlightResponse {
	var filteredFlights []services.FlightResponse
	for _, flight := range flights {
		if flight.Airplane.Name == name {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}

func filterByDeptime(flights []services.FlightResponse, depTimeFrom, dapTimeTo time.Time) []services.FlightResponse {
	var filteredFlights []services.FlightResponse
	for _, flight := range flights {
		if flight.DepTime.After(depTimeFrom) && flight.DepTime.Before(dapTimeTo) {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}

func filterByRemainingSeats(flights []services.FlightResponse, remainingSeats int32) []services.FlightResponse {
	var filteredFlights []services.FlightResponse
	for _, flight := range flights {
		if flight.RemainingSeats >= remainingSeats {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
