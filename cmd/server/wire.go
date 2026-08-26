//go:build wireinject

package main

import (
	"context"
	"io"
	"os"

	"github.com/google/wire"

	"github.com/rubentanahara/user_service/internal/logging"
	"github.com/rubentanahara/user_service/internal/user"
)

func provideLogOutput() io.Writer {
	return os.Stdout
}

func InitializeServer(ctx context.Context) (*Server, func(), error) {
	wire.Build(
		provideMongoURI,
		provideMongoDBName,
		providePort,
		provideMaxPoolSize,
		provideJWTSecret,
		provideLogOutput,
		logging.NewLogger,
		provideMongoClient,
		provideUsersCollection,
		user.NewMongoRepository,
		user.NewService,
		provideTokenIssuer,
		user.NewHandler,
		NewServer,
	)
	return nil, nil, nil
}
