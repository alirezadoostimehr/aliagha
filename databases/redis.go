package databases

import (
	"aliagha/config"
	"fmt"

	"github.com/go-redis/redis"
)

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0, // use default database
	})

	// Test the connection
	pong, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Errorf("error connecting to Redis: %s", err))
	}

	fmt.Println("connected to Redis database: ", pong)
	return client, nil
}
