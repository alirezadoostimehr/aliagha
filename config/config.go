package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Redis          Redis
	Database       Database
	Server         Server
	PaymentGateway PaymentGateway
	MockAPI        MockAPI
	Security       Security
	JWT            JWT
	Zarinpal       Zarinpal
}

type Redis struct {
	Host     string
	Port     int
	Password string
	TTL      time.Duration
}

type Database struct {
	Driver   string
	Host     string
	Port     int
	DB       string
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
	URL     string
	Timeout time.Duration
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

type JWT struct {
	SecretKey string
	ExpiresIn time.Duration
}

type Zarinpal struct {
	MerchantId  string
	CallbackUrl string
	SandBox     bool
}

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
		TTL:      viper.GetDuration("redis.TTL"),
	}
	database := &Database{
		Driver:   viper.GetString("database.driver"),
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		Charset:  viper.GetString("database.chaset"),
		DB:       viper.GetString("database.db"),
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
		URL:     viper.GetString("mock_api.url"),
		Timeout: viper.GetDuration("mock_api.timeout"),
	}

	security := &Security{
		SecretKey:           viper.GetString("security.secret_key"),
		EncryptionAlgorithm: viper.GetString("security.encryption_algorithm"),
	}

	expiresIn := viper.GetDuration("jwt.expires_in")

	if err != nil {
		panic(err)
	}

	jwt := &JWT{
		SecretKey: viper.GetString("jwt.secret_key"),
		ExpiresIn: expiresIn,
	}

	zarinpal := &Zarinpal{
		MerchantId:  viper.GetString("zarinpal.merchant_id"),
		CallbackUrl: viper.GetString("zarinpal.callback_url"),
		SandBox:     viper.GetBool("zarinpal.sand_box"),
	}

	return &Config{
		Redis:          *redis,
		Database:       *database,
		Server:         *server,
		PaymentGateway: *paymentGateway,
		MockAPI:        *mockAPI,
		Security:       *security,
		JWT:            *jwt,
		Zarinpal:       *zarinpal,
	}, nil
}
