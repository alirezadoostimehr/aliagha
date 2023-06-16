package handler

import (
	"aliagha/config"
	"aliagha/models"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	DB  *gorm.DB
	jwt *config.Jwt
}

type CustomClaims struct {
	UserID    int32  `json:"user_id"`
	Cellphone string `json:"cellphone"`
	jwt.StandardClaims
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID int32  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

func validateLoginRequest(req *LoginRequest) error {
	validate := validator.New()
	return validate.Struct(req)
}

func (u *User) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "")
	}
	if err := validateLoginRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var user models.User
	err := u.DB.Where("email	 = ?", req.Email).First(&user).Error
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(c.Response(), "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(c.Response(), "Database error", http.StatusInternalServerError)
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(c.Response(), "Invalid credentials", http.StatusUnauthorized)
	}

	expirationTime := u.jwt.ExpiresAt
	claims := &CustomClaims{
		UserID:    user.ID,
		Cellphone: user.Cellphone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "Aliagha",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.jwt.SecretKey)
	if err != nil {
		http.Error(c.Response(), "Token generation failed", http.StatusInternalServerError)
	}

	response := LoginResponse{
		Token:  tokenString,
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	}

	c.Response().Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Response()).Encode(response)
	return nil
}
