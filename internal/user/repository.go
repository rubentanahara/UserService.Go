package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound  = errors.New("user not found")
	ErrDuplicate = errors.New("username or email already taken")
)

//go:generate mockgen -source=repository.go -destination=repository_mock_test.go -package=user
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id string, u *User) (*User, error)
	UpdatePassword(ctx context.Context, id, hash string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int64) ([]*User, error)
}

type mongoRepository struct {
	coll *mongo.Collection
}

func NewMongoRepository(coll *mongo.Collection) Repository {
	return &mongoRepository{coll: coll}
}

// EnsureIndexes creates the unique constraints this repository relies on.
// Call once at startup.
func EnsureIndexes(ctx context.Context, coll *mongo.Collection) error {
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		return fmt.Errorf("ensure user indexes: %w", err)
	}
	return nil
}

func (r *mongoRepository) Create(ctx context.Context, u *User) error {
	res, err := r.coll.InsertOne(ctx, u)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicate
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	u.ID = res.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoRepository) GetByID(ctx context.Context, id string) (*User, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, err)
	}

	var u User
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user %s: %w", id, err)
	}
	return &u, nil
}

func (r *mongoRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email %s: %w", email, err)
	}
	return &u, nil
}

func (r *mongoRepository) Update(ctx context.Context, id string, u *User) (*User, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, err)
	}

	update := bson.M{"$set": bson.M{
		"username":   u.Username,
		"email":      u.Email,
		"updated_at": u.UpdatedAt,
	}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated User
	err = r.coll.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&updated)
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrDuplicate
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user %s: %w", id, err)
	}
	return &updated, nil
}

func (r *mongoRepository) UpdatePassword(ctx context.Context, id, hash string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, err)
	}

	update := bson.M{"$set": bson.M{
		"password":   hash,
		"updated_at": time.Now().Unix(),
	}}
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return fmt.Errorf("update password %s: %w", id, err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mongoRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, err)
	}

	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mongoRepository) List(ctx context.Context, limit, offset int64) ([]*User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(limit).SetSkip(offset)
	cur, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	var users []*User
	if err := cur.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return users, nil
}
