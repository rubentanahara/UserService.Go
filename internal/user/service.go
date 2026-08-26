package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const maxListLimit = 100

var ErrInvalidCredentials = errors.New("invalid credentials")

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package=user
type Service interface {
	Create(ctx context.Context, username, email, password string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, id, username, email string) (*User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int64) ([]*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error
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
		Username:  strings.ToLower(username),
		Email:     strings.ToLower(email),
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
	patch := &User{Username: strings.ToLower(username), Email: strings.ToLower(email), UpdatedAt: time.Now().Unix()}
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

func (s *service) List(ctx context.Context, limit, offset int64) ([]*User, error) {
	if limit <= 0 || limit > maxListLimit {
		limit = maxListLimit
	}
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, strings.ToLower(email))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *service) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.repo.UpdatePassword(ctx, id, string(hash)); err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	return nil
}
