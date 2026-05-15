package config

import "os"

var (
	APIKey     = os.Getenv("API_KEY")
	DBPassword = os.Getenv("DB_PASSWORD")
)
