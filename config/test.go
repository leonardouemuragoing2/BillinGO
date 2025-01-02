// Package config configures the Application using .env file, environment variables
// and other necessary configs
package config

// TestConfig configures the application to run int TEST mode
var TestConfig = &Config{
	GinPort:                 "5000",
	SecretKey:               "super_secret_key",
	LogLevel:                "DEBUG",
	DBName:                  "billingo_test",
	DBHost:                  "localhost",
	DBPort:                  "5432",
	DBUser:                  "going2",
	DBPassword:              "going2",
	CompressionIntervalDays: 1,
}
