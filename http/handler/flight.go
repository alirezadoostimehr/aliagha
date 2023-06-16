package handler

import (
	"aliagha/models"
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"fmt"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type Flight struct {
	Redis *redis.Client
}

type GetFlightRequest struct { // Add validation and params
	ID            int
	DepartureCity string `json:"departure_city"`
	ArrivalCity   string `json:"arrival_city"`
	date          string
}

func (f *Flight) GetFlightsHandler(c echo.Context) error {
	origin := c.QueryParam("origin")
	dest := c.QueryParam("destination")
	dateStr := c.QueryParam("date")
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
	}

	cacheKey := fmt.Sprintf("%s-%s-%s", origin, dest, dateStr)
	cacheResult, err := f.Redis.Get(cacheKey).Result()
	if err == redis.Nil {
		// Cache miss, get data from API
		apiResult, err := f.getFlightsFromAPI(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get flights from API"})
		}

		// Get filter parameters from query params
		//airline := c.QueryParam("airline")
		//aircraftType := c.QueryParam("aircraft_type")
		depTimeStr := c.QueryParam("dep_time")

		var filteredFlights []models.Flight
		for _, flight := range apiResult {
			//if airline != "" && flight.Airline != airline {
			//	continue
			//}
			//if aircraftType != "" && flight.Airplane.Name != aircraftType {
			//	continue
			//}
			if depTimeStr != "" {
				depTime, err := time.Parse("15:04", depTimeStr)
				if err != nil {
					return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid departure time format"})
				}
				if flight.DepTime.Hour() != depTime.Hour() || flight.DepTime.Minute() != depTime.Minute() {
					continue
				}
			}
			filteredFlights = append(filteredFlights, flight)
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

	var flights []models.Flight
	err = json.Unmarshal([]byte(cacheResult), &flights)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal cache result"})
	}

	return c.JSON(http.StatusOK, flights)
}
func (f *Flight) getFlightsFromAPI(c echo.Context) ([]models.Flight, error) {
	var freq GetFlightRequest
	origin := c.QueryParam("origin")
	dest := c.QueryParam("destination")
	dateStr := c.QueryParam("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
	}
	if err := c.Bind(&freq); err != nil {
		return nil, c.JSON(http.StatusBadRequest, "")
	}
	url := fmt.Sprintf("https://github.com/kianakholousi/Flight-Data-API?origin=%s&destination=%s&date=%s", origin, dest, date)

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
