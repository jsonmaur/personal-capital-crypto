package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	CFG_REDIS_URL string

	CFG_PC_USERNAME       string
	CFG_PC_PASSWORD       string
	CFG_PC_CRYPTO_ACCOUNT string

	CFG_XBT string
	CFG_XDG string
)

func init() {
	if err := godotenv.Load(); err == nil {
		log.Print("Reading environment variables from `.env`")
	}

	CFG_REDIS_URL = GetEnv("REDIS_URL", "redis://localhost:6379")

	CFG_PC_USERNAME = GetEnv("PC_USERNAME", "")
	CFG_PC_PASSWORD = GetEnv("PC_PASSWORD", "")
	CFG_PC_CRYPTO_ACCOUNT = GetEnv("PC_CRYPTO_ACCOUNT", "")

	CFG_XBT = GetEnv("XBT", "")
	CFG_XDG = GetEnv("XDG", "")
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
