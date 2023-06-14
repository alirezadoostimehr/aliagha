package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Redis          Redis
	Database       Database
	Server         Server
	PaymentGateway PaymentGateway
	MockAPI        MockAPI
	Security       Security
}
type Redis struct {
	Host     string
	Port     int
	Password string
}
type Database struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	Username string
	Password string
	Charset  string
}
type Server struct {
	Address string
	Port    int
}

type PaymentGateway struct {
	URL       string
	APIKey    string
	APISecret string
}

type MockAPI struct {
	URL        string
	AuthKey    string
	AuthSecret string
}

type Security struct {
	SecretKey           string
	EncryptionAlgorithm string
}
type Params struct {
	FilePath string
	FileName string
	FileType string
}

// Load all  uration values from YAML file
func Init(param Params) (*Config, error) {
	viper.SetConfigType(param.FileType)
	viper.AddConfigPath(param.FilePath)
	// viper.SetConfigFile("./config/config.yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}
	redis := &Redis{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
	}
	database := &Database{
		Driver:   viper.GetString("database.driver"),
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		Charset:  viper.GetString("database.chaset"),
		Name:     viper.GetString("database.name"),
	}
	server := &Server{
		Address: viper.GetString("server.address"),
		Port:    viper.GetInt("server.port"),
	}

	paymentGateway := &PaymentGateway{
		URL:       viper.GetString("payment_gateway.url"),
		APIKey:    viper.GetString("payment_gateway.api_key"),
		APISecret: viper.GetString("payment_gateway.api_secret"),
	}

	mockAPI := &MockAPI{
		URL:        viper.GetString("mock_api.url"),
		AuthKey:    viper.GetString("mock_api.auth_key"),
		AuthSecret: viper.GetString("mock_api.auth_secret"),
	}

	security := &Security{
		SecretKey:           viper.GetString("security.secret_key"),
		EncryptionAlgorithm: viper.GetString("security.encryption_algorithm"),
	}
	return &Config{
		Redis:          *redis,
		Database:       *database,
		Server:         *server,
		PaymentGateway: *paymentGateway,
		MockAPI:        *mockAPI,
		Security:       *security,
	}, nil
}
