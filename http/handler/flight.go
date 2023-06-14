package handler

import (
	"aliagha/models"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"fmt"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type Flight struct {
	Redis    *redis.Client
	Flight   *models.Flight
	Airplane *models.Airplane
}

func (f *Flight) GetFlightsHandler(c echo.Context) error {
	origin := c.QueryParam("origin")
	dest := c.QueryParam("destination")
	dateStr := c.QueryParam("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", origin, dest, dateStr)
	cacheResult, err := f.Redis.Get(cacheKey).Result()
	if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := f.getFlightsFromAPI(origin, dest, date)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flights from API"})
		}
		// Store result in cache
		jsonData, err := json.Marshal(apiResult)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
		}
		err = f.Redis.Set(cacheKey, jsonData, 1*time.Minute).Err()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store result in cache"})
		}
		cacheResult = string(jsonData)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get result from cache"})
	}

	var flights []f.Flight
	err = json.Unmarshal([]byte(cacheResult), &flights)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}

	return c.JSON(http.StatusOK, flights)
}

func (f *Flight) FilterFlightsHandler(c echo.Context) error {
	origin := c.QueryParam("origin")
	dest := c.QueryParam("destination")
	dateStr := c.QueryParam("date")
	startDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
	}
	endDate := startDate.AddDate(0, 0, 1)

	// Get filter parameters from query params
	airline := c.QueryParam("airline")
	aircraftType := c.QueryParam("aircraft_type")
	depTimeStr := c.QueryParam("dep_time")
	emptyCapacity, _ := strconv.Atoi(c.QueryParam("Empty_capacity"))

	cacheKey := fmt.Sprintf("%s-%s-%s-%s-%s-%d", origin, dest, dateStr)
	cacheResult, err := f.Redis.Get(cacheKey).Result()
	if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := f.getFlightsFromAPI(origin, dest, startDate)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flights from API"})
		}
		var filteredFlights []f.Flight
		for _, flight := range apiResult {

			if flight.DepTime.After(startDate) && flight.DepTime.Before(endDate) {
				if airline != "" && flight.Airplane.Airline != airline {
					continue
				}
				if aircraftType != "" && flight.Airplane.Type != aircraftType {
					continue
				}
				if depTimeStr != "" {
					depTime, err := time.Parse("15:04", depTimeStr)
					if err != nil {
						return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid departure time format"})
					}
					if flight.DepTime.Hour() != depTime.Hour() || flight.DepTime.Minute() != depTime.Minute() {
						continue
					}
				}
				if emptyCapacity > 0 && flight.Airplane.Empty_capacity < emptyCapacity {
					continue
				}
				filteredFlights = append(filteredFlights, flight)
			}
		}

		// Sort results based on query params
		sortBy := c.QueryParam("sort_by")
		if sortBy != "" {
			sortOrder := c.QueryParam("sort_order")
			if sortOrder == "" {
				sortOrder = "asc"
			}
			switch sortBy {
			case "price":
				if sortOrder == "asc" {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].Price < filteredFlights[j].Price
					})
				} else {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].Price > filteredFlights[j].Price
					})
				}
			case "dep_time":
				if sortOrder == "asc" {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].DepTime.Before(filteredFlights[j].DepTime)
					})
				} else {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].DepTime.After(filteredFlights[j].DepTime)
					})
				}
			case "duration":
				if sortOrder == "asc" {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].ArrTime.Sub(filteredFlights[i].DepTime) < filteredFlights[j].ArrTime.Sub(filteredFlights[j].DepTime)
					})
				} else {
					sort.Slice(filteredFlights, func(i, j int) bool {
						return filteredFlights[i].ArrTime.Sub(filteredFlights[i].DepTime) > filteredFlights[j].ArrTime.Sub(filteredFlights[j].DepTime)
					})
				}
			default:
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sort_by parameter"})
			}
		}

		// Store result in cache
		jsonData, err := json.Marshal(filteredFlights)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
		}
		err = f.Redis.Set(cacheKey, jsonData, 1*time.Minute).Err()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store result in cache"})
		}
		cacheResult = string(jsonData)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get result from cache"})
	}

	var flights []Flight
	err = json.Unmarshal([]byte(cacheResult), &flights)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}

	return c.JSON(http.StatusOK, flights)
}
func (f *Flight) FindFlightByID(c echo.Context) error {
	id := c.Param("ID")
	cacheKey := fmt.Sprintf("flight-%s", id)
	cacheResult, err := f.Redis.Get(cacheKey).Result()
	if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := f.getFlightFromAPI(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flight from API"})
		}
		// Store result in cache
		jsonData, err := json.Marshal(apiResult)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal API result"})
		}
		err = f.Redis.Set(cacheKey, jsonData, 1*time.Minute).Err()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store result in cache"})
		}
		cacheResult = string(jsonData)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get result from cache"})
	}

	var flight Flight
	err = json.Unmarshal([]byte(cacheResult), &flight)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}

	return c.JSON(http.StatusOK, flight)
}
func (f *Flight) getFlightsFromAPI(origin, dest string, date time.Time) ([]Flight, error) {
	// TODO: API request to get flights
	// For each flight, include the airplane information?
	var flights []f.Flight
	var airplanes []f.Airplane
	for _, flight := range apiResult {
		airplane, err := f.getAirplaneByIDFromAPI(flight.AirplaneID)
		if err != nil {
			return nil, err
		}
		flights = append(flight)
		airplanes = append(airplanes, airplane)

	}
	return flights, nil
}
func (f *Flight) getFlightFromAPI(id string) (Flight, error) {
	// TODO: send API request to get flight by ID
	return Flight{}, nil
}
func (f *Flight) getAirplaneByIDFromAPI(id string) (*models.Airplane, error) {
	// TODO: API request to get airplane by ID Return the airplane object
	return nil, nil
}
func FlightRoutes(e *echo.Echo, redisClient *redis.Client) {
	f := &Flight{Redis: redisClient, Flight: &models.Flight{}}
	e.GET("/flights", f.GetFlightsHandler)
	e.GET("/flights/search", f.FilterFlightsHandler)
	e.GET("/flights/:ID", f.FindFlightByID)
}
