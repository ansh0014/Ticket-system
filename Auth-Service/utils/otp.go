package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/ansh0014/auth/config"
)

func GenerateOTP(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	otp := ""
	for _, b := range bytes {
		otp += fmt.Sprintf("%d", b%10)
	}
	return otp, nil
}

func StoreOTP(email, otp string, expiry time.Duration) error {
	return config.RedisClient.Set(context.Background(), "otp:"+email, otp, expiry).Err()
}

func GetOTP(email string) (string, error) {
	return config.RedisClient.Get(context.Background(), "otp:"+email).Result()
}

func DeleteOTP(email string) error {
	return config.RedisClient.Del(context.Background(), "otp:"+email).Err()
}
