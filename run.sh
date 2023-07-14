#!/bin/bash

# Set the path to the Go binary
GO_BIN="/usr/local/go/bin/go"

# Set the path to the project directory
PROJECT_DIR="$PWD"

cleanup() {
    rm -f "$PROJECT_DIR/mycommand"
}

# Set up trap to call cleanup function on INT (Ctrl+C)
trap cleanup INT

# Build the project
cd "$PROJECT_DIR" || exit
"$GO_BIN" build -o mycommand .

# Take down existing migrations 
./mycommand migrate -c config -a down -f migrations

# Run the migrate command
./mycommand migrate -c config -a up -f migrations

# Run the serve command
./mycommand serve -c config

