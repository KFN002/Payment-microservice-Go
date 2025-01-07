package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config Общая конфигурация
type Config struct {
	Server   Server   `yaml:"server" env-prefix:"SERVER_"`
	Postgres Postgres `yaml:"postgres" env-prefix:"POSTGRES_"`
	Redis    Redis    `yaml:"redis" env-prefix:"REDIS_"`
	Forex    Forex    `yaml:"forex" env-prefix:"FOREX_"`
	Yoomoney Yoomoney `yaml:"yoomoney" env-prefix:"YOOMONEY_"`
}

// Server конфигурация сервера
type Server struct {
	Port int `yaml:"Port" env:"PORT"`
}

// Postgres конфигурация бд
type Postgres struct {
	Host     string `yaml:"Host" env:"HOST"`
	Port     int    `yaml:"Port" env:"PORT"`
	SSLMode  string `yaml:"SSLMode" env:"SSL_MODE"`
	DB       string `yaml:"DB" env:"DB"`
	User     string `yaml:"User" env:"USER"`
	Password string `yaml:"Password" env:"PASSWORD"`
}

// Redis конфигурация редиски
type Redis struct {
	URL string `yaml:"URL" env:"URL"`
}

// Forex конфигурация форекса для api
type Forex struct {
	Key string `yaml:"Key" env:"KEY"`
}

// Yoomoney юмани для оплаты
type Yoomoney struct {
	Token    string `yaml:"Token" env:"TOKEN"`
	ClientID string `yaml:"ClientID" env:"CLIENT_ID"`
	Receiver int    `yaml:"Receiver" env:"RECEIVER"`
}

// LoadConfig загрузка конфигурации
func LoadConfig() (*Config, error) {
	configPath, exists := os.LookupEnv("CONFIG_PATH")
	if !exists {
		return nil, errors.New("Missing CONFIG_PATH env variable")
	}
	var config Config
	var err error
	if configPath == "environment" {
		err = cleanenv.ReadEnv(&config)
	} else {
		err = cleanenv.ReadConfig(configPath, &config)
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to process config: %v", err)
	}
	return &config, nil
}
