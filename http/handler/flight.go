package handler

import (
	"aliagha/config"

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
	Config    *config.Config
}
type FlightModuleAPI struct {
	ID         int32
	DepCityID  int32
	DepCity    City
	ArrCityID  int32
	ArrCity    City
	DepTime    time.Time
	ArrTime    time.Time
	Date       time.Time
	AirplaneID int32
	Airplane   Airplane
	Airline    string
	Price      int32
	CxlSitID   int32
	LeftSeat   int32
}
type City struct {
	ID   int32
	Name string
}
type Airplane struct {
	ID   int32
	Name string
}
type GetRequest struct {
	DepartureCity City      `query:"departure_city" validate:"required"`
	ArrivalCity   City      `query:"arrival_city" validate:"required"`
	Date          time.Time `query:"date" validate:"required,datetime"`
	Airline       string    `query:"airline"`
	Name          string    `query:"name"`
	Deptime       time.Time `query:"departure_time"` /*validate:"time"`*/
	SortBy        string    `query:"sort_by"`
	SortOrder     string    `query:"sort_order"`
	EmptySeats    int32     `query:"left_seat"`
}

var timeout = 10 * time.Second

func (f *Flight) Get(ctx echo.Context) error {
	TTL := f.Config.Redis.TTL
	var req GetRequest
	var flights []FlightModuleAPI
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "")
	}

	if err := f.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// req.Date = req.Deptime.Format("2006-01-02")
	cacheKey := fmt.Sprintf("%s-%s-%s", req.DepartureCity.Name, req.ArrivalCity.Name, req.Date.Format("2003-02-01"))
	cacheResult, err := f.Redis.Get(cacheKey).Bytes()

	err = json.Unmarshal(cacheResult, &flights)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	if err != nil && err != redis.Nil {
		return ctx.JSON(http.StatusInternalServerError, "Inetrnal Server Error")
	} else if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := getFlightsFromAPI(req.DepartureCity, req.ArrivalCity, req.Date)
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
	}

	// Filter
	var filteredFlights []FlightModuleAPI
	if err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid airline ID"})
	}

	if req.Airline != "" {
		filteredFlights = append(filteredFlights, filterByAirline(flights, req.Airline)...)
	}

	if req.Name != "" {
		filteredFlights = append(filteredFlights, filterByName(flights, req.Name)...)
	}

	if req.Deptime.Format("2003-02-01") != "" {

		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid departure time format"})
		}

		filteredFlights = append(filteredFlights, filterByDeptime(flights, req.Deptime)...)
	}
	if req.EmptySeats > 0 {
		filteredFlights = append(filteredFlights, filterByLeftSeat(flights, req.EmptySeats)...)
	}

	// Sort
	if req.SortBy != "" {
		flights, err := sortFlight(flights, req.SortBy, req.SortOrder)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sort argument"})
		}

		filteredFlights = append(filteredFlights, flights...)
	}

	// Return filtered results
	jsonData, err := json.Marshal(filteredFlights)
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

func getFlightsFromAPI(depCity, arrCity City, date time.Time) ([]FlightModuleAPI, error) {

	url := fmt.Sprintf("https://github.com/kianakholousi/Flight-Data-API?departure_city=%s&arrival_city%s&date=%s", depCity.Name, arrCity.Name, date.Format("2003-02-01"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResult []FlightModuleAPI
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return nil, err
	}

	return apiResult, nil
}
func getFlightFromAPI(c echo.Context) (FlightModuleAPI, error) {
	id := c.QueryParam("id")
	url := fmt.Sprintf("https://github.com/kianakholousi/Flight-Data-API/%s", id)
	var flight FlightModuleAPI
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return flight, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return flight, err
	}
	defer resp.Body.Close()

	var apiResult FlightModuleAPI
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return flight, err
	}

	return apiResult, nil
}
func sortFlight(flights []FlightModuleAPI, sortBy, sortOrder string) ([]FlightModuleAPI, error) {
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
func filterByAirline(flights []FlightModuleAPI, filter string) []FlightModuleAPI {
	var filteredFlights []FlightModuleAPI
	for _, flight := range flights {
		if flight.Airline == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByName(flights []FlightModuleAPI, filter string) []FlightModuleAPI {
	var filteredFlights []FlightModuleAPI
	for _, flight := range flights {
		if flight.Airplane.Name == filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByDeptime(flights []FlightModuleAPI, filter time.Time) []FlightModuleAPI {
	var filteredFlights []FlightModuleAPI
	for _, flight := range flights {
		if flight.DepTime.Hour() == filter.Hour() && flight.DepTime.Minute() == filter.Minute() {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
func filterByLeftSeat(flights []FlightModuleAPI, filter int32) []FlightModuleAPI {
	var filteredFlights []FlightModuleAPI
	for _, flight := range flights {
		if flight.LeftSeat >= filter {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights
}
