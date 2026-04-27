package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBTimeZone        string
	JWTSecret         string
	JWTExpirationTime int
	RedisHost         string
	RedisPassword     string
	ServerPort        string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}
	if jwt, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION_HOURS")); err == nil {
		return &Config{
			ServerPort:        os.Getenv("PORT"),
			DBHost:            os.Getenv("DB_HOST"),
			DBPort:            os.Getenv("DB_PORT"),
			DBUser:            os.Getenv("DB_USER"),
			DBPassword:        os.Getenv("DB_PASSWORD"),
			DBName:            os.Getenv("DB_NAME"),
			DBSSLMode:         os.Getenv("DB_SSLMODE"),
			JWTSecret:         os.Getenv("JWT_SECRET"),
			DBTimeZone:        os.Getenv("DB_TIMEZONE"),
			JWTExpirationTime: jwt,
			RedisHost:         os.Getenv("REDIS_ADDR"),
			RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		}, nil
	}
	return nil, fmt.Errorf("config.LoadConfig: %w", err)
}
