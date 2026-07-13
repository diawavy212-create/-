package config

import "os"

type Config struct {
	HTTPAddr           string
	MySQLDSN           string
	RedisAddr          string
	RedisPassword      string
	CASEndpoint        string
	MessageGatewayURL  string
	WeChatAppID        string
	WeChatAppSecret    string
	AdminTokenAudience string
	AuthTokenSecret    string
	AdminLoginPassword string
	CASServiceURL      string
	DevAuthEnabled     bool
}

func Load() Config {
	devAuthEnabled := env("DEV_AUTH_ENABLED", "true") == "true"
	adminLoginPassword := env("ADMIN_LOGIN_PASSWORD", "")
	if adminLoginPassword == "" && devAuthEnabled {
		adminLoginPassword = "admin123456"
	}
	return Config{
		HTTPAddr:           env("HTTP_ADDR", ":8090"),
		MySQLDSN:           env("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/teacher_platform?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:          env("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:      env("REDIS_PASSWORD", ""),
		CASEndpoint:        env("CAS_ENDPOINT", "https://cas.example.edu"),
		MessageGatewayURL:  env("MESSAGE_GATEWAY_URL", "https://message.example.edu"),
		WeChatAppID:        env("WECHAT_APP_ID", ""),
		WeChatAppSecret:    env("WECHAT_APP_SECRET", ""),
		AdminTokenAudience: env("ADMIN_TOKEN_AUDIENCE", "teacher-platform-admin"),
		AuthTokenSecret:    env("AUTH_TOKEN_SECRET", "dev-change-me-before-production"),
		AdminLoginPassword: adminLoginPassword,
		CASServiceURL:      env("CAS_SERVICE_URL", "http://127.0.0.1:5173"),
		DevAuthEnabled:     devAuthEnabled,
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
