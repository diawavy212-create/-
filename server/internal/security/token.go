package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID int64
	Role   string
	Expiry time.Time
}

func SignToken(secret string, userID int64, role string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("token secret is empty")
	}
	expiry := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d.%s.%d", userID, role, expiry)
	signature := sign(secret, payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature, nil
}

func VerifyToken(secret string, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, errors.New("invalid token format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, err
	}
	payload := string(payloadBytes)
	if !hmac.Equal([]byte(sign(secret, payload)), []byte(parts[1])) {
		return Claims{}, errors.New("invalid token signature")
	}

	fields := strings.Split(payload, ".")
	if len(fields) != 3 {
		return Claims{}, errors.New("invalid token payload")
	}
	userID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return Claims{}, err
	}
	expiryUnix, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return Claims{}, err
	}
	expiry := time.Unix(expiryUnix, 0)
	if time.Now().After(expiry) {
		return Claims{}, errors.New("token expired")
	}

	return Claims{UserID: userID, Role: fields[1], Expiry: expiry}, nil
}

func sign(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
