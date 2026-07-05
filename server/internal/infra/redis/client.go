package redis

import "teacher-platform/server/internal/config"

type Client struct {
	Addr     string
	Password string
}

func New(cfg config.Config) Client {
	return Client{Addr: cfg.RedisAddr, Password: cfg.RedisPassword}
}
