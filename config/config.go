package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Configs struct {
	DBConfig   DatabaseConfig
	Sconfig    ServerConfig
	PGconfig   PaymentGatewayConfig
	MAPIConfig MockAPIConfig
	SeConfig   SecurityConfig
}
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
func LoadConfig(f *os.File) (*Configs, error) {

	// viper.AddConfigPath(".")
	viper.SetConfigType("ymal")
	// viper.SetConfigFile(fName)
	err := viper.ReadConfig(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s", err)
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
	configs := Configs{
		DBConfig:   *dbConfig,
		Sconfig:    *serverConfig,
		PGconfig:   *paymentGatewayConfig,
		MAPIConfig: *mockAPIConfig,
		SeConfig:   *securityConfig,
	}
	return &configs, nil
}
