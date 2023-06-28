package services

import (
	"aliagha/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/eapache/go-resiliency/breaker"
)

var ErrForbidden = errors.New("forbidden access")

type APIMockClient struct {
	Client  *http.Client
	Breaker *breaker.Breaker
	BaseURL string
	APIKey  string
}

type FlightResponse struct {
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

func (c *APIMockClient) GetFlights(depCity, arrCity string, date time.Time) ([]FlightResponse, error) {
	url := c.BaseURL + "/flights" + fmt.Sprintf("=%s&arrival_city%s&date=%s", depCity, arrCity, date.Format("2003-02-01"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)

	var resp []FlightResponse
	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		req = req.WithContext(ctx)

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_get_flights: request failed, error: %w", err)
		}

		defer response.Body.Close()

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("apimock_get_flights: read response body failed, error: %w", err)
		}

		switch response.StatusCode {
		case http.StatusOK:
			err = json.Unmarshal(responseBody, &resp)
			if err != nil {
				return fmt.Errorf("apimock_get_flights: parse response body failed, response: %s, error: %w", responseBody, err)
			}

			return nil
		case http.StatusForbidden:
			return ErrForbidden
		default:
			return fmt.Errorf("apimock_get_flights: unhandeled response, status: %d, response: %s", response.StatusCode, responseBody)
		}
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
