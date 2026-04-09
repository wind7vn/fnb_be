package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	Port    string `mapstructure:"PORT"`
	Env     string `mapstructure:"ENV"`
	AppName string `mapstructure:"APP_NAME"`

	// Database
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	DBSSLMode      string `mapstructure:"DB_SSLMODE"`
	DBMaxIdleConns int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `mapstructure:"DB_MAX_OPEN_CONNS"`

	// Redis
	RedisURL      string `mapstructure:"REDIS_URL"`
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// Auth
	JWTSecret            string `mapstructure:"JWT_SECRET"`
	JWTExpireMinutes     int    `mapstructure:"JWT_EXPIRE_MINUTES"`
	JWTRefreshExpireDays int    `mapstructure:"JWT_REFRESH_EXPIRE_DAYS"`

	// PUSH
	FirebaseServiceAccountPath string `mapstructure:"FIREBASE_SERVICE_ACCOUNT_PATH"`
}

var AppConfig Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found or error reading .env file, relying on environment variables only.")
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
