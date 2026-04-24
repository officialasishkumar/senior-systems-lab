package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr           string
	TCPAddr            string
	UDPAddr            string
	LogLevel           string
	QueueCapacity      int
	DeadLetterCapacity int
	Workers            int
	ShutdownTimeout    time.Duration
}

func FromEnv() Config {
	return Config{
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		TCPAddr:            getenv("TCP_ADDR", ":9090"),
		UDPAddr:            getenv("UDP_ADDR", ":9091"),
		LogLevel:           getenv("LOG_LEVEL", "info"),
		QueueCapacity:      getenvInt("QUEUE_CAPACITY", 1024),
		DeadLetterCapacity: getenvInt("DLQ_CAPACITY", 128),
		Workers:            getenvInt("WORKERS", 4),
		ShutdownTimeout:    time.Duration(getenvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second,
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
