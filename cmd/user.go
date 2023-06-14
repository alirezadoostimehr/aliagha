package cmd

import (
	"aliagha/models"

	"github.com/spf13/cobra"

	"fmt"
	"strconv"
)

// Define the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Perform user operations",
}

// Define subcommands for user operations
var (
	indexCmd = &cobra.Command{
		Use:   "index",
		Short: "Get all users",
		Run:   indexUser,
	}

	getCmd = &cobra.Command{
		Use:   "get <id>",
		Short: "Get a user by ID",
		Args:  cobra.ExactArgs(1),
		Run:   getUser,
	}

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run:   createUser,
	}

	// Flags for the createCmd
	createNameFlag     string
	createPasswordFlag string
	createMobileFlag   string
	createEmailFlag    string
)

func init() {
	// Add subcommands to the user command
	userCmd.AddCommand(indexCmd, getCmd, createCmd)

	// Add flags to the createCmd
	createCmd.Flags().StringVarP(&createNameFlag, "name", "n", "", "User name")
	createCmd.Flags().StringVarP(&createPasswordFlag, "password", "p", "", "User password")
	createCmd.Flags().StringVarP(&createMobileFlag, "mobile", "m", "", "User mobile number")
	createCmd.Flags().StringVarP(&createEmailFlag, "email", "e", "", "User email address")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("password")
	createCmd.MarkFlagRequired("mobile")
	createCmd.MarkFlagRequired("email")
}

// AddUserCommands adds the user command to the root command
func AddUserCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(userCmd)
}

// Handler functions for the subcommands
func indexUser(cmd *cobra.Command, args []string) {
	var users []models.User
	// Fetch users from the database or any other data source
	// For now, let's assume it returns an empty slice
	// users := []models.User{}
	fmt.Println("Fetching all users...")
	fmt.Println(users)
}

func getUser(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	user := &models.User{ID: int32(id)}
	// Fetch user from the database or any other data source
	// For now, let's assume it returns a user with the given ID
	// user := &models.User{ID: int32(id), Name: "John Doe", ...}
	fmt.Println("Fetching user with ID:", id)
	fmt.Println(user)
}

func createUser(cmd *cobra.Command, args []string) {
	// Read user data from the flags
	user := models.User{
		Name:     createNameFlag,
		Password: createPasswordFlag,
		Mobile:   createMobileFlag,
		Email:    createEmailFlag,
	}

	// Save the user to the database or any other data source
	// For now, let's assume it was saved successfully
	fmt.Println("Creating user...")
	fmt.Println(user)
}

