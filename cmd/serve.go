package cmd

import (
	"aliagha/config"
	"aliagha/databases"
	"aliagha/http/handler"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startServer() {
	//todo: get config path as a flag
	cfg, err := config.Init("config.yaml")
	if err != nil {
		panic(err)
	}

	redis, err := databases.InitRedis(cfg)
	if err != nil {
		panic(err)
	}

	database, err := databases.InitDB(cfg)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	flight := handler.Flight{Redis: redis}
	e.GET("/flight", flight.Get)

	user := handler.User{DB: database}

	e.Start("localhost:3030")
}
