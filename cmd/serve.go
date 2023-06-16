package cmd

import (
	"aliagha/config"
	"aliagha/database"
	"aliagha/http/handler"
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
	//todo: get config path as a flag
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

	flight := handler.Flight{Redis: redis}
	e.GET("/flights", flight.Get)

	user := handler.User{DB: db, JWT: &cfg.JWT, Validator: vldt}
	e.POST("/user/login", user.Login)

	e.Start("localhost:3030")
}
