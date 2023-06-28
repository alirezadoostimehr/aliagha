package handler

import (
	"aliagha/config"
	"aliagha/helpers"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"golang.org/x/crypto/bcrypt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserTestSuite struct {
	suite.Suite
	sqlMock   sqlmock.Sqlmock
	e         *echo.Echo
	user      *User
	mockToken string
}

func (suite *UserTestSuite) SetupSuite() {
	mockDB, sqlMock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}))

	if err != nil {
		log.Fatal(err)
	}

	suite.sqlMock = sqlMock
	vldt := validator.New()
	suite.user = &User{DB: db, JWT: &config.JWT{
		SecretKey: "secretkey",
		ExpiresIn: 3600,
	}, Validator: vldt}
	suite.mockToken = "testToken"
	suite.e = echo.New()
}

func (suite *UserTestSuite) CallHandler(requestBody string, endPoint string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(http.MethodPost, endPoint, strings.NewReader(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	c := suite.e.NewContext(req, res)

	var err error
	if endPoint == "/user/login" {
		err = suite.user.Login(c)
	} else {
		err = suite.user.Register(c)
	}

	if err != nil {
		return res, err
	}

	return res, nil
}

func (suite *UserTestSuite) TestUserLogin_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusOK
	expectedResponse := `{"token":"` + suite.mockToken + `"}`

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	email := "test@yahoo.com"
	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "cellphone", "email", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "hashedPassword", "1234567890", email, time.Now(), time.Now()))

	monkey.Patch(bcrypt.CompareHashAndPassword, func(hashedPassword, password []byte) error {
		return nil
	})
	defer monkey.Unpatch(bcrypt.CompareHashAndPassword)

	monkey.Patch(helpers.GenerateJwtToken, func(userID int32, cellphone string, jwt *config.JWT) (string, error) {
		return suite.mockToken, nil
	})
	defer monkey.Unpatch(helpers.GenerateJwtToken)

	res, err := suite.CallHandler(`{"email":"test@example.com","password":"1234567"}`, "/user/login")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func (suite *UserTestSuite) TestUserLogin_Validation_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest

	res, err := suite.CallHandler(`{"email":"test","password":"1234567"}`, "/user/login")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *UserTestSuite) TestUserLogin_UserFinding_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusUnauthorized

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WithArgs("test@yahoo.com").
		WillReturnError(sql.ErrNoRows)

	res, err := suite.CallHandler(`{"email":"test@yahoo.com","password":"1234567"}`, "/user/login")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *UserTestSuite) TestUserLogin_CompareHash_Failure() {
	require := suite.Require()
	expectedResponse := `"Invalid Credentials"`
	expectedStatusCode := http.StatusUnauthorized

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WithArgs("test@yahoo.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "cellphone", "email", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "hashedPassword", "1234567890", "test@yahoo.com", time.Now(), time.Now()))

	monkey.Patch(bcrypt.CompareHashAndPassword, func(hashedPassword, password []byte) error {
		return errors.New("")
	})
	defer monkey.Unpatch(bcrypt.CompareHashAndPassword)

	res, err := suite.CallHandler(`{"email":"test@yahoo.com","password":"1234567"}`, "/user/login")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func (suite *UserTestSuite) TestUserLogin_GenerateJWTToken_Failure() {
	require := suite.Require()
	expectedResponse := `"Internal Server Error"`
	expectedStatusCode := http.StatusInternalServerError

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WithArgs("test@yahoo.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "cellphone", "email", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "hashedPassword", "1234567890", "test@yahoo.com", time.Now(), time.Now()))

	monkey.Patch(bcrypt.CompareHashAndPassword, func(hashedPassword, password []byte) error {
		return nil
	})
	defer monkey.Unpatch(bcrypt.CompareHashAndPassword)

	monkey.Patch(helpers.GenerateJwtToken, func(userID int32, cellphone string, jwt *config.JWT) (string, error) {
		return "", errors.New("")
	})
	defer monkey.Unpatch(helpers.GenerateJwtToken)

	res, err := suite.CallHandler(`{"email":"test@yahoo.com","password":"1234567"}`, "/user/login")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func (suite *UserTestSuite) TestUserRegister_Success() {
	require := suite.Require()
	expectedStatusCode := http.StatusCreated

	expectedResponse := `{
		"message": "User created successfully",
		"token": "` + suite.mockToken + `"
	}`

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WithArgs("test@yahoo.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).
			AddRow(1, "test@yahoo.com"))

	monkey.Patch(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
		return []byte("test"), nil
	})
	defer monkey.Unpatch(bcrypt.GenerateFromPassword)

	suite.sqlMock.ExpectBegin()
	suite.sqlMock.ExpectExec("^INSERT INTO `users` VALUES (.+)").
		WithArgs(1, "test", "09123456789", "test@yahoo.com", "hashedPassword", time.Now(), time.Now()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	suite.sqlMock.ExpectCommit()

	monkey.Patch(helpers.GenerateJwtToken, func(userID int32, cellphone string, jwt *config.JWT) (string, error) {
		return suite.mockToken, nil
	})
	defer monkey.Unpatch(helpers.GenerateJwtToken)

	res, err := suite.CallHandler(
		`{"email":"test@yahoo.com","cellphone":"09123456789","name":"matin khalili", "password":"1234567"}`,
		"/user/register")
	fmt.Println(res.Body.String())
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.JSONEq(expectedResponse, res.Body.String())
}

func (suite *UserTestSuite) TestUserRegister_Validation_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusBadRequest

	res, err := suite.CallHandler(
		`{"email":"test@yahoo.com","cellphone":"09123456789","name":"matin khalili", "password":"123"}`,
		"/user/register")
	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
}

func (suite *UserTestSuite) TestUserRegister_UserExist_Failure() {
	require := suite.Require()
	expectedStatusCode := http.StatusUnprocessableEntity
	expectedResponse := `"Email already exists"`

	monkey.Patch(suite.user.Validator.Struct, func(s interface{}) error {
		return nil
	})
	defer monkey.Unpatch(suite.user.Validator.Struct)

	suite.sqlMock.ExpectQuery("^SELECT (.+) FROM `users` WHERE email = (.+)").
		WithArgs("test@yahoo.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "cellphone", "email", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "hashedPassword", "1234567890", "test@yahoo.com", time.Now(), time.Now()))

	res, err := suite.CallHandler(
		`{"email":"test@yahoo.com","cellphone":"09123456789","name":"matin khalili", "password":"1234567"}`,
		"/user/register")

	require.NoError(err)
	require.Equal(expectedStatusCode, res.Code)
	require.Equal(expectedResponse, strings.TrimSpace(res.Body.String()))
}

func TestCreateUser(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
