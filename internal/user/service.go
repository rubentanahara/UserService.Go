package user

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, username, email, password string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, id, username, email string) (*User, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().Unix()
	u := &User{
		Username:  username,
		Email:     email,
		Password:  string(hash),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	return u, nil
}

func (s *service) Update(ctx context.Context, id, username, email string) (*User, error) {
	patch := &User{Username: username, Email: email, UpdatedAt: time.Now().Unix()}
	u, err := s.repo.Update(ctx, id, patch)
	if err != nil {
		return nil, fmt.Errorf("update user %s: %w", id, err)
	}
	return u, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}
	return nil
}
