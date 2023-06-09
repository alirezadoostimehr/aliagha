package databases

import (
	"fmt"

	config "aliagha/config"

	"github.com/go-redis/redis"
)

func init_redis() {
	var configObj *config.Configs
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", configObj.DBConfig.Host, configObj.DBConfig.Port),
		Password: configObj.DBConfig.Password,
		DB:       0, // use default database
	})

	// Test the connection
	pong, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Errorf("Error connecting to Redis: %s", err))
	}
	fmt.Println("connected to Redis database: ", pong)
}
