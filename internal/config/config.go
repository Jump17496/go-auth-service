package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL string `mapstructure:"database_url"`
	JWTSecret   string `mapstructure:"jwt_secret"`
	Port        string `mapstructure:"port"`
}

func Load() *Config {
	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set default values
	viper.SetDefault("database_url", "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable")
	viper.SetDefault("jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("port", "8080")

	// Read environment variables (will override config file values)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.BindEnv("database_url", "DATABASE_URL")
	viper.BindEnv("jwt_secret", "JWT_SECRET")
	viper.BindEnv("port", "PORT")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults and environment variables")
		} else {
			log.Printf("Error reading config file: %v\n", err)
		}
	} else {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &cfg
}
