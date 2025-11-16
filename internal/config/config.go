package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment       string
	Server            ServerConfig
	Database          DatabaseConfig
	Logger            LoggerConfig
	EmailNotification EmailConfig
	MQTTBroker        MQTTBrokerConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL     string
	Debug   bool
	ConPool int
}

type LoggerConfig struct {
	Level  string
	Format string
}

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type MQTTBrokerConfig struct {
	URL              string
	Username         string
	Password         string
	ClientID         string
	PayloadSeparator string
}

func Load() (*Config, error) {
	env := os.Getenv("ENV")
	env = os.Getenv("GO_ENV")
	if "" == env {
		env = "development"
	}

	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env." + env)
	_ = godotenv.Load(".env.defaults")

	// todo: consider adding validation of loaded envs
	config := &Config{
		Environment: env,
		Server: ServerConfig{
			Port: os.Getenv("PORT"),
		},
		Database: DatabaseConfig{
			URL:     os.Getenv("DB_URL"),
			Debug:   getBoolEnv("DB_DEBUG"),
			ConPool: getIntEnv("DB_CON_POOL", 5),
		},
		Logger: LoggerConfig{
			Level:  os.Getenv("LOG_LEVEL"),
			Format: os.Getenv("LOG_FORMAT"),
		},
		EmailNotification: EmailConfig{
			Host:     os.Getenv("EMAIL_NOTIFICATION_HOST"),
			Port:     os.Getenv("EMAIL_NOTIFICATION_PORT"),
			Username: os.Getenv("EMAIL_NOTIFICATION_USERNAME"),
			Password: os.Getenv("EMAIL_NOTIFICATION_PASSWORD"),
			From:     os.Getenv("EMAIL_NOTIFICATION_FROM"),
		},
		MQTTBroker: MQTTBrokerConfig{
			URL:              os.Getenv("MQTT_BROKER_URL"),
			Username:         os.Getenv("MQTT_BROKER_USERNAME"),
			Password:         os.Getenv("MQTT_BROKER_PASSWORD"),
			ClientID:         os.Getenv("MQTT_BROKER_CLIENT_ID"),
			PayloadSeparator: os.Getenv("MQTT_BROKER_PAYLOAD_SEPARATOR"),
		},
	}

	return config, nil
}

func getIntEnv(key string, def int) int {
	val := os.Getenv(key)
	if "" == val {
		return def
	}

	atoi, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(atoi)
}

func getBoolEnv(key string) bool {
	return os.Getenv(key) == "true"
}

func getDurationEnv(key string) time.Duration {
	val := getIntEnv(key, 0)

	return time.Duration(val) * time.Second
}

func getBytesEnv(key string) []byte {
	return []byte(os.Getenv(key))
}
