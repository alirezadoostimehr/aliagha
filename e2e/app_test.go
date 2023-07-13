package e2e

import (
	"aliagha/services"
	utl "aliagha/utils"
	"bou.ke/monkey"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os/exec"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	baseUrl = "http://127.0.0.1:3030"
)

var (
	client = http.Client{}
)

type loginResponse struct {
	Token string `json:"token"`
}

func callHandler(method, path, body, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, baseUrl+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func TestApp(t *testing.T) {
	cmd := exec.Command("./run.sh")
	cmd.Dir = "/home/alireza/go/src/aliagha"
	err := cmd.Start()
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)

	t.Run("Testing Register", func(t *testing.T) {
		reqPath := "/user/register"

		reqBody := `{"name": "ali", "cellphone": "1234567890"}`
		resp, err := callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest, "Register reacts abnormally to invalid input")
		t.Log("Register reacts normally to invalid input")

		reqBody = `{"name": "ali", "cellphone": "1234567890", "email":"test@yahoo.com", "password":"1234567"}`
		resp, err = callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusCreated, "Register reacts abnormally to valid input")
		t.Log("Register reacts normally to valid input")

		reqBody = `{"name": "ali", "cellphone": "1234567890", "email":"test@yahoo.com", "password":"1234567"}`
		resp, err = callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest, "Register reacts abnormally to repeated input")
		t.Log("Register reacts normally to repeated input")

		reqBody = `{"name": "ali", "cellphone": "0234567890", "email":"test2@yahoo.com", "password":"1234567"}`
		resp, err = callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusCreated, "Register reacts abnormally to second valid input")
		t.Log("Register reacts normally to second valid input")

	})

	var loginToken loginResponse

	t.Run("Testing Login", func(t *testing.T) {
		reqPath := "/user/login"

		reqBody := `{"name": "ali", "cellphone": "1234567890"}`
		resp, err := callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest, "Login reacts abnormally to invalid input")
		t.Log("Login reacts normally to imperfect input")

		reqBody = `{"email":"test@yahoo.com", "password":"wrongPassword"}`
		resp, err = callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusUnauthorized, "Login reacts abnormally to imperfect input")
		t.Log("Login reacts normally to imperfect input")

		reqPath = "/user/login"
		reqBody = `{"email":"test@yahoo.com", "password":"1234567"}`
		resp, err = callHandler("POST", reqPath, reqBody, "")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK, "Login reacts abnormally to valid input")
		t.Log("Login reacts normally to valid input")

		responseBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err, "Login returns problematic body")

		err = json.Unmarshal(responseBody, &loginToken)
		assert.NoError(t, err, "Login returns problematic body")
		t.Log("Login returns fine body")
	})

	t.Run("Making passengers", func(t *testing.T) {
		reqPath := "/passengers"

		reqBody := `{"name": "ali", "cellphone": "1234567890"}`
		resp, err := callHandler("POST", reqPath, reqBody, loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "Passenger reacts abnormally to invalid input")
		t.Log("Login reacts normally to imperfect input")

		reqBody = `{"name": "ali", "national_code": "1213548980", "birth_date": "2003-02-01"}`
		resp, err = callHandler("POST", reqPath, reqBody, loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Passenger reacts abnormally to valid input")
		t.Log("Login reacts normally to valid input")

		reqBody = `{"name": "ali", "national_code": "1213548980", "birth_date": "2003-02-01"}`
		resp, err = callHandler("POST", reqPath, reqBody, loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "Passenger reacts abnormally to repeated input")
		t.Log("Login reacts normally to repeated input")

		reqBody = `{"name": "ali", "national_code": "1213548981", "birth_date": "2003-02-01"}`
		resp, err = callHandler("POST", reqPath, reqBody, loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Passenger reacts abnormally to second valid input")
		t.Log("Login reacts normally to second valid input")

		reqBody = `{"name": "ali", "national_code": "1213548982", "birth_date": "2003-02-01"}`
		resp, err = callHandler("POST", reqPath, reqBody, loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Passenger reacts abnormally to third valid input")
		t.Log("Login reacts normally to third valid input")
	})

	passengers := make(map[string]interface{})
	t.Run("Getting Passengers", func(t *testing.T) {
		reqPath := "/passengers"
		resp, err := callHandler("GET", reqPath, "", loginToken.Token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Passenger reacts abnormally to valid request for getting")
		t.Log("Passenger reacts normally to valid request for getting")

		responseBody, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err, "Passenger(get) returns problematic body")

		err = json.Unmarshal(responseBody, &passengers)
		assert.NoError(t, err, "Passenger(get) returns problematic body")
		t.Log("Passenger(get) returns fine body")
	})

	t.Run("Requesting for flight info", func(t *testing.T) {
		monkey.PatchInstanceMethod(reflect.TypeOf(&services.APIMockClient{}), "GetFlights", func(_ *services.APIMockClient, depCity string, arrCity string, date string) ([]services.FlightResponse, error) {
			res := make([]services.FlightResponse, 0, 1)
			depTime, _ := utl.ParseDate(date)
			res = append(res, services.FlightResponse{
				ID: 1,
				DepCity: services.City{
					ID:   1,
					Name: depCity,
				},
				ArrCity: services.City{
					ID:   2,
					Name: arrCity,
				},
				DepTime: depTime,
				ArrTime: depTime.Add(1 * time.Hour),
				Airplane: services.Airplane{
					ID:   1,
					Name: "Boeing777",
				},
				Airline:        "Emirates",
				Price:          1000,
				CxlSitID:       1,
				RemainingSeats: 10,
			})
			return res, nil
		})
		reqPath := "/flights?departure_city=Athens&arrival_city=London&date=2020-04-11"
		resp, err := callHandler("GET", reqPath, "", "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Get flight reacts abnormally to valid request")
		t.Log("Get flight reacts normally to valid request for getting")
		monkey.UnpatchAll()

	})
	err = cmd.Process.Signal(syscall.SIGINT)
	assert.NoError(t, err)
}
