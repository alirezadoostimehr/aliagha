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
		migrateCobra()
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
var Config, MigrationFolder string

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().VarP(&Action, "action", "a", `action to perform: "up" or "down"`)
	migrateCmd.Flags().StringVarP(&Config, "config", "c", "", "path to custom configuration file in YAML format")
	migrateCmd.Flags().StringVarP(&MigrationFolder, "folder", "f", "", "path to migration folder")
}

func migrateCobra() {
	cfg, err := config.Init(config.Params{FilePath: Config, FileType: "yaml"})
	if err != nil {
		panic(err)
	}

	err = MigrateDB(
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		strconv.Itoa(cfg.Database.Port),
		cfg.Database.Name,
		Action.String())
	if err != nil && err.Error() != "no change" {
		panic(err)
	}

}

func MigrateDB(username, password, host, port, dbname, action string) error {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		username, password, host, port)

	dbStart, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	_, err = dbStart.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", "aliagha"))
	if err != nil {
		return err
	}

	dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		username, password, host, port, dbname)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", MigrationFolder),
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	if action == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}
	return err
}
