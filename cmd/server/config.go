package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type MongoURI string
type MongoDBName string
type Port string
type MaxPoolSize uint64
type JWTSecret string

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func provideMongoURI() MongoURI {
	return MongoURI(getenv("MONGO_URI", "mongodb://localhost:27017"))
}

func provideMongoDBName() MongoDBName {
	return MongoDBName(getenv("MONGO_DB", "user_service"))
}

func providePort() Port {
	return Port(getenv("PORT", "5001"))
}

func provideMaxPoolSize() (MaxPoolSize, error) {
	v := os.Getenv("MONGO_MAX_POOL_SIZE")
	if v == "" {
		return defaultMaxPoolSize, nil
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid MONGO_MAX_POOL_SIZE: %w", err)
	}
	return MaxPoolSize(n), nil
}

func provideJWTSecret() (JWTSecret, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is required")
	}
	return JWTSecret(secret), nil
}
