package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	GinPort                 string `config:"PORT"`
	SecretKey               string `config:"SECRET_KEY"`
	LogLevel                string `config:"LOG_LEVEL"`
	DBName                  string `config:"POSTGRES_NAME"`
	DBHost                  string `config:"POSTGRES_HOST"`
	DBPort                  string `config:"POSTGRES_PORT"`
	DBUser                  string `config:"POSTGRES_USER"`
	DBPassword              string `config:"POSTGRES_PWD"`
	CompressionIntervalDays int    `config:"COMPRESSION_INTERVAL_DAYS"`
}

// loadDotEnv load the configuration inside the application taking care
// if the running OS is Windows making the necessary adaptations for this case.
// There is a precedence order: exported env var > .env file >  default config
func loadDotEnv(environmentFile string) {
	if environmentFile == "" {
		environmentFile = ".env"
	}
	envFile, hasPath := os.LookupEnv("BILLING_DIR")
	if hasPath {
		log.Infof("Looking for Settings File at %v...\n", envFile)
		godotenv.Load(envFile + environmentFile)
	} else if runtime.GOOS == "windows" {
		executable, _ := os.Executable()
		executable = filepath.FromSlash(executable)
		directory := filepath.Dir(executable)
		log.Infof("Looking for Settings File at %v...\n", directory)
		godotenv.Load(filepath.Join(directory, environmentFile))
	} else {
		godotenv.Load(environmentFile)
	}
}

// populateConfig check environment variables
// to set to the current configuration
func populateConfig(config *Config) {
	loadDotEnv("")
	value := reflect.ValueOf(config)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	for i := 0; i < value.NumField(); i++ {
		tag := value.Type().Field(i).Tag.Get("config")
		defaultVal := value.Type().Field(i).Tag.Get("default")
		if defaultVal != "" {
			if value.Field(i).Kind() == reflect.Int {
				valInt, err := strconv.Atoi(defaultVal)
				if err != nil {
					panic(fmt.Errorf("%v env is not a valid int: %w", defaultVal, err))
				}
				value.Field(i).SetInt(int64(valInt))
			} else if value.Field(i).Kind() == reflect.Bool {
				valBool, err := strconv.ParseBool(defaultVal)
				if err != nil {
					panic(fmt.Errorf("%v env is not a valid bool: %w", defaultVal, err))
				}
				value.Field(i).SetBool(valBool)
			} else {
				value.Field(i).SetString(defaultVal)
			}
		}
		if tag != "" {
			if env, exists := os.LookupEnv(tag); exists {
				if value.Field(i).Kind() == reflect.Int {
					valInt, err := strconv.Atoi(env)
					if err != nil {
						panic(fmt.Errorf("%v env is not a valid int: %w", env, err))
					}
					value.Field(i).SetInt(int64(valInt))
				} else if value.Field(i).Kind() == reflect.Bool {
					valBool, err := strconv.ParseBool(env)
					if err != nil {
						panic(fmt.Errorf("%v env is not a valid bool: %w", env, err))
					}
					value.Field(i).SetBool(valBool)
				} else {
					value.Field(i).SetString(env)
				}
			}
		}
	}
}

// LoadConfig return the Configuration based on the current environment
func LoadConfig() *Config {
	env := os.Getenv("ENVIRONMENT")
	var conf *Config
	switch env {
	case "DEVELOPMENT":
		conf = DevConfig
	case "TESTING":
		conf = TestConfig
	case "PRODUCTION":
		conf = ProdConfig
	default:
		conf = ProdConfig
	}
	populateConfig(conf)
	return conf
}
