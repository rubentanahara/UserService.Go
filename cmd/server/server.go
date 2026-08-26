package main

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/rubentanahara/user_service/internal/auth"
	"github.com/rubentanahara/user_service/internal/user"
)

type Server struct {
	Handler *user.Handler
	Logger  *slog.Logger
	Mongo   *mongo.Client
	Port    Port
}

func NewServer(h *user.Handler, logger *slog.Logger, client *mongo.Client, port Port) *Server {
	return &Server{Handler: h, Logger: logger, Mongo: client, Port: port}
}

func provideTokenIssuer(secret JWTSecret) *auth.TokenIssuer {
	return auth.NewTokenIssuer(string(secret))
}

func provideMongoClient(ctx context.Context, uri MongoURI, poolSize MaxPoolSize) (*mongo.Client, func(), error) {
	client, err := mongo.Connect(options.Client().ApplyURI(string(uri)).SetMaxPoolSize(uint64(poolSize)))
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo: %w", err)
	}
	cleanup := func() {
		_ = client.Disconnect(context.Background())
	}
	return client, cleanup, nil
}

func provideUsersCollection(ctx context.Context, client *mongo.Client, db MongoDBName) (*mongo.Collection, error) {
	coll := client.Database(string(db)).Collection("users")
	if err := user.EnsureIndexes(ctx, coll); err != nil {
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}
	return coll, nil
}
