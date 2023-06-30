package database

import (
	"aliagha/config"
	"fmt"
	"log"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

func InitRedis(redisConfig *config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       0, // use default database
	})

	pong, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Errorf("error connecting to Redis: %s", err))
	}

	fmt.Println("connected to Redis database: ", pong)
	return client, nil
}

func NewRedisMock() (*miniredis.Miniredis, *redis.Client) {
	server, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{Addr: server.Addr()})

	return server, client
}
