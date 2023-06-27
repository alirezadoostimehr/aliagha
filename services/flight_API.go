package services

import (
	"aliagha/config"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type FlightModuleAPI struct {
	ID             int32
	DepCity        City
	ArrCity        City
	DepTime        time.Time
	ArrTime        time.Time
	Date           time.Time
	Airplane       Airplane
	Airline        string
	Price          int32
	CxlSitID       int32
	RemainingSeats int32
}
type City struct {
	ID   int32
	Name string
}
type Airplane struct {
	ID   int32
	Name string
}

var c *config.Config
var timeout = c.MockAPI.Request_timeout

func GetFlightsFromAPI(depCity, arrCity City, date time.Time) ([]FlightModuleAPI, error) {

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
func GetFlightFromAPI(c echo.Context) (FlightModuleAPI, error) {
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
