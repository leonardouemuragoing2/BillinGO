// Package config configures the Application using .env file, environment variables
// and other necessary configs
package config

// ProdConfig configures the application to run int PRODUCTION mode
var ProdConfig = &Config{
	GinPort:                 "5000",
	SecretKey:               "",
	LogLevel:                "INFO",
	DBName:                  "",
	DBHost:                  "",
	DBPort:                  "",
	DBUser:                  "",
	DBPassword:              "",
	CompressionIntervalDays: 90,
}
