//go:build integration

package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoRepository_CaseInsensitiveDuplicate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongo container: %v", err)
	}
	t.Cleanup(func() {
		if err := mongoContainer.Terminate(context.Background()); err != nil {
			t.Errorf("terminate mongo container: %v", err)
		}
	})

	uri, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Errorf("disconnect mongo: %v", err)
		}
	})

	coll := client.Database("user_service_test").Collection("users")
	if err := EnsureIndexes(ctx, coll); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}

	svc := NewService(NewMongoRepository(coll))

	if _, err := svc.Create(ctx, "jane", "Jane@x.com", "password123"); err != nil {
		t.Fatalf("create first user: %v", err)
	}

	_, err = svc.Create(ctx, "janedoe", "jane@x.com", "password123")
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("create second user: err = %v, want %v", err, ErrDuplicate)
	}
}
