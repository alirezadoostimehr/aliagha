/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"aliagha/config"
	"errors"
	"fmt"
	"strconv"

	db "aliagha/database"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database",
	Long: `Migrate database up or down

This command migrates the database up or down to a specific schema version. The direction of the migration is determined by the value of the --action flag, which can be set to "up" or "down".
	
It is recommended to run this command before starting the application to ensure that the necessary tables and columns are available.
	
You must specify a custom configuration file in YAML format using the --config flag. By default, this command will not run without a configuration file.
	
Usage:
	migrate --config [path] --action [up/down]
	
Flags:
	-a, --action string   Action to perform: "up" or "down" (required)
	-c, --config string   Path to custom configuration file in YAML format (required)`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

type myEnum string

func (e *myEnum) String() string {
	return string(*e)
}

func (e *myEnum) Set(v string) error {
	switch v {
	case "up", "down":
		*e = myEnum(v)
		return nil
	default:
		return errors.New(`must be either "up" or "down"`)
	}
}

func (e *myEnum) Type() string {
	return "string"
}

var Action myEnum
var Config string

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().VarP(&Action, "action", "a", `action to perform: "up" or "down"`)
	migrateCmd.Flags().StringVarP(&Config, "config", "c", "", "path to custom configuration file in YAML format")
}

func migrate() {
	cfg, err := config.Init(config.Params{FilePath: Config, FileType: "yaml"})
	if err != nil {
		panic(err)
	}
	username := cfg.Database.Username
	password := cfg.Database.Password
	host := cfg.Database.Host
	port := cfg.Database.Port
	dbname := cfg.Database.Name
	address := cfg.Database.MigrationAddress

	fmt.Println(db.MigrateDB(username, password, host, strconv.Itoa(port), dbname, address, Action.String()))

}
