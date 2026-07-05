package cas

import "teacher-platform/server/internal/config"

type Client struct {
	Endpoint string
}

func New(cfg config.Config) Client {
	return Client{Endpoint: cfg.CASEndpoint}
}
