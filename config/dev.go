// Package config configures the Application using .env file, environment variables
// and other necessary configs
package config

// DevConfig configures the application to run int DEVELOPMENT mode
var DevConfig = &Config{
	GinPort:                 "5000",
	SecretKey:               "super_secret_key",
	LogLevel:                "DEBUG",
	DBName:                  "billingo",
	DBHost:                  "localhost",
	DBPort:                  "5432",
	DBUser:                  "going2",
	DBPassword:              "going2",
	CompressionIntervalDays: 1,
}
