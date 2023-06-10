package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Configs struct {
	RConfig    Redis
	DBConfig   DatabaseConfig
	Sconfig    ServerConfig
	PGconfig   PaymentGatewayConfig
	MAPIConfig MockAPIConfig
	SeConfig   SecurityConfig
}
type Redis struct {
	Host     string
	Port     int
	Password string
}
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	Username string
	Password string
	Charset  string
	// charset  utf8mb4
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
func LoadConfig(filePath string) (*Configs, error) {

	viper.SetConfigType("yaml")
	viper.AddConfigPath(filePath)
	// viper.SetConfigFile(fName)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
	}
	redisConfig := &Redis{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
	}
	dbConfig := &DatabaseConfig{
		Driver:   viper.GetString("database.driver"),
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		Charset:  viper.GetString("database.chaset"),
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
	return &Configs{
		RConfig:    *redisConfig,
		DBConfig:   *dbConfig,
		Sconfig:    *serverConfig,
		PGconfig:   *paymentGatewayConfig,
		MAPIConfig: *mockAPIConfig,
		SeConfig:   *securityConfig,
	}, nil
}
