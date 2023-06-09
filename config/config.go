package config

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Password string
}

type ServerConfig struct {
	Address string
	Port    int
}

type PaymentGatewayConfig struct {
	URL       string
	APIKey    string
	APISecret string
}

type MockAPIConfig struct {
	URL        string
	AuthKey    string
	AuthSecret string
}

type SecurityConfig struct {
	SecretKey           string
	EncryptionAlgorithm string
}

// Load all configuration values from YAML file
func LoadConfig() (*DatabaseConfig, *ServerConfig, *PaymentGatewayConfig, *MockAPIConfig, *SecurityConfig, error) {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to read config file: %s", err)
	}
	dbConfig := &DatabaseConfig{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Password: viper.GetString("database.password"),
	}
	serverConfig := &ServerConfig{
		Address: viper.GetString("server.address"),
		Port:    viper.GetInt("server.port"),
	}

	paymentGatewayConfig := &PaymentGatewayConfig{
		URL:       viper.GetString("payment_gateway.url"),
		APIKey:    viper.GetString("payment_gateway.api_key"),
		APISecret: viper.GetString("payment_gateway.api_secret"),
	}

	mockAPIConfig := &MockAPIConfig{
		URL:        viper.GetString("mock_api.url"),
		AuthKey:    viper.GetString("mock_api.auth_key"),
		AuthSecret: viper.GetString("mock_api.auth_secret"),
	}

	securityConfig := &SecurityConfig{
		SecretKey:           viper.GetString("security.secret_key"),
		EncryptionAlgorithm: viper.GetString("security.encryption_algorithm"),
	}

	return dbConfig, serverConfig, paymentGatewayConfig, mockAPIConfig, securityConfig, nil
}

func main() {

	// Load database config
	dbConfig, _, _, _, _, err := LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config : %s", err))
	}

	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port),
		Password: dbConfig.Password,
		DB:       0, // use default database
	})

	// Test the connection
	pong, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Errorf("Error connecting to Redis: %s", err))
	}
	fmt.Println("connected to Redis database: ", pong)
}
