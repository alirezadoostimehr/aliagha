package handler

import "github.com/go-redis/redis"

type Passenger struct {
	DB *redis.Client //gorm
}
