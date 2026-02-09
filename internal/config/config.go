package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	BindAddress   string
	LogFile       string
	TailscaleOnly bool
}

func Load() *Config {
	cfg := &Config{
		Port:          getEnv("CHAT_PORT", "8022"),
		BindAddress:   getEnv("CHAT_BIND", ""), // Empty means all interfaces
		LogFile:       getEnv("CHAT_LOG_FILE", "secure_chat.log"),
		TailscaleOnly: getEnvBool("CHAT_TAILSCALE_ONLY", true),
	}

	log.Printf("Configuration loaded: Port=%s, Bind=%s, TailscaleOnly=%t",
		cfg.Port, cfg.BindAddress, cfg.TailscaleOnly)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
