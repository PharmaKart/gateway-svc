package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	AuthServiceURL      string
	ProductServiceURL   string
	OrderServiceURL     string
	PaymentServiceURL   string
	ReminderServiceURL  string
	StripeWebhookSecret string
}

func LoadConfig() *Config {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	return &Config{
		Port:                getEnv("PORT", "8080"),
		AuthServiceURL:      getEnv("AUTH_SERVICE_URL", "localhost:50051"),
		ProductServiceURL:   getEnv("PRODUCT_SERVICE_URL", "localhost:50052"),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "localhost:50053"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "localhost:50054"),
		ReminderServiceURL:  getEnv("REMINDER_SERVICE_URL", "localhost:50055"),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", "whsec_your_stripe_webhook_secret"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
