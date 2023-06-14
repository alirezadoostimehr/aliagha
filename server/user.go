package server

import (
	"aliagha/models"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

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

