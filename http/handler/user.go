package handler

import (
	"aliagha/config"
	"aliagha/helpers"
	"aliagha/models"
	"database/sql"
	"time"

	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	DB        *gorm.DB
	JWT       *config.JWT
	Validator *validator.Validate
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (u *User) Login(ctx echo.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := u.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	var user models.User
	err := u.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusUnauthorized, "Invalid Credentials")
		} else {
			return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, "Invalid Credentials")
	}

	token, err := helpers.GenerateJwtToken(user.ID, user.Cellphone, u.JWT)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusOK, LoginResponse{
		Token: token,
	})
}

type RegisterRequest struct {
	Name      string `json:"name" validate:"required,min=3,max=100"`
	Cellphone string `json:"cellphone" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6,max=20"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func (u *User) Register(ctx echo.Context) error {
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := u.Validator.Struct(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	var user models.User
	err := u.DB.Model(&models.User{}).Where("email = ?", req.Email).First(&user).Error

	if err == nil {
		return ctx.JSON(http.StatusUnprocessableEntity, "Email already exists")
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(http.StatusInternalServerError, "Internal server error")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal server error")
	}

	var token string
	err = u.DB.Transaction(func(tx *gorm.DB) error {
		user = models.User{
			Name:      req.Name,
			Cellphone: req.Cellphone,
			Email:     req.Email,
			Password:  string(hashedPassword),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		token, err = helpers.GenerateJwtToken(user.ID, user.Cellphone, u.JWT)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.JSON(http.StatusCreated, RegisterResponse{
		Message: "User created successfully",
		Token:   token,
	})
}
