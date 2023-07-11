package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/breaker"
)

type APIMockClient struct {
	Client  *http.Client
	Breaker *breaker.Breaker
	BaseURL string
	Timeout time.Duration
}

type FlightResponse struct {
	ID             int32     `json:"id"`
	DepCity        City      `json:"dep_city"`
	ArrCity        City      `json:"arr_city"`
	DepTime        time.Time `json:"dep_time"`
	ArrTime        time.Time `json:"arr_time"`
	Airplane       Airplane  `json:"airplane"`
	Airline        string    `json:"airline"`
	Price          int32     `json:"price"`
	CxlSitID       int32     `json:"cxl_sit_id"`
	RemainingSeats int32     `json:"remaining_seats"`
}

type City struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type Airplane struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (c *APIMockClient) GetFlights(depCity, arrCity, date string) ([]FlightResponse, error) {
	url := c.BaseURL + "/flights" + fmt.Sprintf("?departure_city=%s&arrival_city=%s&date=%s", depCity, arrCity, date)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	var resp []FlightResponse
	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_get_flights: request failed, error: %w", err)
		}

		defer response.Body.Close()

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("apimock_get_flights: read response body failed, error: %w", err)
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_get_flights: unhandeled response, status: %d, response: %s", response.StatusCode, responseBody)
		}

		err = json.Unmarshal(responseBody, &resp)
		if err != nil {
			return fmt.Errorf("apimock_get_flights: parse response body failed, response: %s, error: %w", responseBody, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *APIMockClient) Reserve(flightId, cnt int32) error {
	url := c.BaseURL + "/flights/reserve"

	reqBody := fmt.Sprintf(`{"flight_id": %d, "count": %d}`, flightId, cnt)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_post_reserve: request failed, error: %v", err.Error())
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_post_reserve: request failed, error: %v", err.Error())
		}

		return nil
	})

	return err
}

func (c *APIMockClient) Cancel(flightId, cnt int32) error {
	url := c.BaseURL + "/flights/cancel"

	reqBody := fmt.Sprintf(`{"flight_id": %d, "count": %d}`, flightId, cnt)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_post_cancel: request failed, error: %v", err.Error())
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_post_cancel: request failed, error: %v", err.Error())
		}

		return nil
	})

	return err
}

type FlightInfoResponse struct {
	ID               int32     `json:"id"`
	DepCity          City      `json:"dep_city"`
	ArrCity          City      `json:"arr_city"`
	DepTime          time.Time `json:"dep_time"`
	ArrTime          time.Time `json:"arr_time"`
	Airplane         Airplane  `json:"airplane"`
	Airline          string    `json:"airline"`
	Price            int32     `json:"price"`
	CxlSitID         int32     `json:"cxl_sit_id"`
	RemainingSeats   int32     `json:"remaining_seats"`
	FlightClass      string    `json:"flight_class"`
	BaggageAllowance string    `json:"baggage_allowance"`
	MealService      string    `json:"meal_service"`
	Gate             string    `json:"gate"`
}

func (c *APIMockClient) GetFlightInfo(flightId int32) (FlightInfoResponse, error) {
	url := c.BaseURL + fmt.Sprintf("/flight?flight_id=%d", flightId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return FlightInfoResponse{}, err
	}

	var flight FlightInfoResponse
	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_get_flight_info: request failed, error: %v", err.Error())
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("apimock_get_flight_info: reading response failed, error: %v", err.Error())
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_get_flight_info: request failed, error: %v", err.Error())
		}

		if err := json.Unmarshal(responseBody, &flight); err != nil {
			return fmt.Errorf("apimock_get_flight_info: parsing response body failed, error: %v", err.Error())
		}

		return nil
	})

	if err != nil {
		return FlightInfoResponse{}, err
	}

	return flight, nil
}

type GetCityResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (c *APIMockClient) GetCities() ([]GetCityResponse, error) {
	url := c.BaseURL + "/cities"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var resp []GetCityResponse
	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_get_cities: request failed, error: %v", err.Error())
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("apimock_get_cities: reading response failed, error: %v", err.Error())
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_get_cities: request failed, error: %v", err.Error())
		}

		if err := json.Unmarshal(responseBody, &resp); err != nil {
			return fmt.Errorf("apimock_get_cities: parsing response body failed, error: %v", err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetAirplaneResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (c *APIMockClient) GetAirplanes() ([]GetAirplaneResponse, error) {
	url := c.BaseURL + "/airplanes"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var resp []GetAirplaneResponse
	err = c.Breaker.Run(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
		defer cancel()

		req = req.WithContext(ctx)

		response, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("apimock_get_airplanes: request failed, error: %v", err.Error())
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("apimock_get_airplanes: reading response failed, error: %v", err.Error())
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("apimock_get_airplanes: request failed, error: %v", err.Error())
		}

		if err := json.Unmarshal(responseBody, &resp); err != nil {
			return fmt.Errorf("apimock_get_airplanes: parsing response body failed, error: %v", err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
