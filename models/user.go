package models


import (
	
	"net/http"
	"strconv"
	"time"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Password  string    `gorm:"column:password;not null" json:"password"`
	Mobile    string    `gorm:"column:mobile;not null" json:"mobile"`
	Email     string    `gorm:"column:email;not null" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func indexUser(c echo.Context) error {
	var user []models.User
	return c.JSON(http.StatusOK, user)
}
func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	user := &models.User{ID: id}
	return c.JSON(http.StatusOK, user)
}
func createUser(c echo.Context) error {
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, user)
}

