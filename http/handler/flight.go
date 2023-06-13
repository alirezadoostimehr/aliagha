package handler

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type Flight struct {
	Redis *redis.Client
}

func (f *Flight) Get(c echo.Context) error {
	return nil
}
