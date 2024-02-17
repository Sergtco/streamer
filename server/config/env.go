package config

import (
	"github.com/joho/godotenv"
)
func init() {
    InitEnv()
}

// Initializes following variables.
//
// CATALOGUE=./catalogue/
//
// HLS=./hls/
//
// DB_PATH=./database.db
func InitEnv() {
    if err := godotenv.Load("./config/envVariables.env"); err != nil {
        panic("Error loading environment variables")
    }
}
