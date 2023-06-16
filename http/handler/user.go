package handler

import (
	"aliagha/config"
	"aliagha/models"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	DB        *gorm.DB
	JWT       *config.JWT
	Validator *validator.Validate
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
	Token string `json:"token"`
}

func (u *User) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := u.Validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var user models.User
	err := u.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, "Invalid Credentials")
		} else {
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid Credentials")
	}

	claims := &CustomClaims{
		UserID:    user.ID,
		Cellphone: user.Cellphone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: u.JWT.ExpiresIn.Unix(),
			Issuer:    "Aliagha",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.JWT.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Token: tokenString,
	})
}
