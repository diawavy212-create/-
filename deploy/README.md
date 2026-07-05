# Production Deployment Assets

This directory contains production templates:

- `env/teacher-platform.env.example`: environment variables for the API service
- `systemd/teacher-platform-api.service`: systemd service for the Go API binary
- `nginx/teacher-platform.conf`: HTTPS reverse proxy and admin static hosting
- `logrotate/teacher-platform`: API log rotation rule

Replace every placeholder domain, password, secret, and certificate path before deployment.
