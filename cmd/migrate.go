package cmd

import (
	"aliagha/config"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database",
	Long: `This command migrates the database to a schema version.

The action to perform is determined by the --action flag, which can be set to "up" or "down".

You must specify a custom configuration file in YAML format using the --config flag. By default, this command will not run without a configuration file.

You must also specify a custom folder path for your migration files using the --folder flag.

It is recommended to run this command before starting the application to ensure that the necessary tables and columns are available.
	
Usage:
	mycommand migrate --config [path] --action [up/down] --folder [path]
	
Flags:
	-a, --action string   Action to perform: "up" or "down" (required)
	-c, --config string   Path to custom configuration file in YAML format (required)
	-f, --folder string   Path to custom folder for migration files (required)
	-h, --help            help for migrate
	
It is recommended to run this command before starting the application to ensure that the necessary tables and columns are available.`,
	Run: func(cmd *cobra.Command, args []string) {
		if action != "up" && action != "down" {
			panic(errors.New("invalid action"))
		}
		
		migrateDB()
	},
}

var action string
var configPath, migrationFolder string

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&action, "action", "a", "", `action to perform: "up" or "down"`)
	migrateCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to custom configuration file in YAML format")
	migrateCmd.Flags().StringVarP(&migrationFolder, "folder", "f", "", "path to migration folder")
}

func migrateDB() {
	cfg, err := config.Init(config.Params{FilePath: configPath, FileType: "yaml"})
	if err != nil {
		panic(err)
	}

	username := cfg.Database.Username
	password := cfg.Database.Password
	host := cfg.Database.Host
	port := strconv.Itoa(cfg.Database.Port)
	dbname := cfg.Database.Name

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		username, password, host, port)

	dbStart, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}

	_, err = dbStart.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		panic(err)
	}

	dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		username, password, host, port, dbname)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationFolder),
		"mysql",
		driver,
	)
	if err != nil {
		panic(err)
	}

	if action == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}

	if err != nil {
		panic(err)
	}
}
