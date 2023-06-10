package databases

import (
	"fmt"

	"github.com/go-redis/redis"
)

func InitRedis(cfg *Config.Config) (*redis.Client, error) {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.DBConfig.Host, cfg.DBConfig.Port),
		Password: cfg.DBConfig.Password,
		DB:       0, // use default database
	})

	// Test the connection
	pong, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Errorf("Error connecting to Redis: %s", err))
	}
	fmt.Println("connected to Redis database: ", pong)
	return client, nil
}
