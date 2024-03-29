package cmd

import (
	"aliagha/config"
	"aliagha/database"
	"aliagha/http/handler"
	"aliagha/http/middleware"
	"aliagha/services"
	"net/http"

	"github.com/eapache/go-resiliency/breaker"
	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

var serveConfigPath string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&serveConfigPath, "config", "c", "", "Path to the YAML configuration file (required)")
	if err := serveCmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}
}

func startServer() {
	cfg, err := config.Init(config.Params{FilePath: serveConfigPath, FileType: "yaml"})
	if err != nil {
		panic(err)
	}

	redis, err := database.InitRedis(&cfg.Redis)
	if err != nil {
		panic(err)
	}

	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		panic(err)
	}

	vldt := validator.New()

	e := echo.New()

	mockClient := services.APIMockClient{
		Client:  &http.Client{},
		Breaker: &breaker.Breaker{},
		BaseURL: cfg.MockAPI.URL,
		Timeout: cfg.MockAPI.Timeout,
	}

	flight := handler.Flight{Redis: redis, Validator: vldt, Config: cfg, APIMock: mockClient}
	// jwtMiddleware := middleware.AuthenticatorMiddleware(cfg.JWT.SecretKey)

	e.GET("/flights", flight.Get)

	user := handler.User{DB: db, JWT: &cfg.JWT, Validator: vldt}
	e.POST("/user/login", user.Login)
	e.POST("/user/register", user.Register)

	passenger := handler.Passenger{DB: db, Validator: vldt}
	e.POST("/passengers", passenger.CreatePassenger, middleware.AuthMiddleware(cfg.JWT.SecretKey))
	e.GET("/passengers", passenger.GetPassengers, middleware.AuthMiddleware(cfg.JWT.SecretKey))

	ticket := handler.Ticket{DB: db}
	e.GET("/tickets", ticket.GetTickets, middleware.AuthMiddleware(cfg.JWT.SecretKey))

	flightReservation := handler.FlightReservation{DB: db, Validator: vldt, APIMock: mockClient}
	e.POST("/flights/reserve", flightReservation.Reserve, middleware.AuthMiddleware(cfg.JWT.SecretKey))

	e.GET(cfg.Zarinpal.CallbackUrl, flightReservation.VerifyPayment)

	err = e.Start("localhost:3030")
	if err != nil {
		panic(err)
	}

}
