package messaging

import "teacher-platform/server/internal/config"

type Client struct {
	GatewayURL string
}

func New(cfg config.Config) Client {
	return Client{GatewayURL: cfg.MessageGatewayURL}
}
