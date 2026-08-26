package user

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		mockErr  error
		wantErr  error
	}{
		{
			name:     "lowercases username and email",
			username: "Jane",
			email:    "Jane@Example.com",
			password: "password123",
		},
		{
			name:     "propagates duplicate",
			username: "jane",
			email:    "jane@example.com",
			password: "password123",
			mockErr:  ErrDuplicate,
			wantErr:  ErrDuplicate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, u *User) error {
					if u.Username != "jane" {
						t.Errorf("username = %q, want lowercased %q", u.Username, "jane")
					}
					if u.Email != "jane@example.com" {
						t.Errorf("email = %q, want lowercased %q", u.Email, "jane@example.com")
					}
					if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(tt.password)); err != nil {
						t.Errorf("stored password does not match hash: %v", err)
					}
					return tt.mockErr
				},
			)

			svc := NewService(repo)
			_, err := svc.Create(context.Background(), tt.username, tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr error
	}{
		{name: "found"},
		{name: "not found", mockErr: ErrNotFound, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			want := &User{Username: "jane"}
			repo.EXPECT().GetByID(gomock.Any(), "1").Return(want, tt.mockErr)

			svc := NewService(repo)
			got, err := svc.GetByID(context.Background(), "1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got != want {
				t.Fatalf("got = %v, want %v", got, want)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr error
	}{
		{name: "updated"},
		{name: "not found", mockErr: ErrNotFound, wantErr: ErrNotFound},
		{name: "duplicate", mockErr: ErrDuplicate, wantErr: ErrDuplicate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().Update(gomock.Any(), "1", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, u *User) (*User, error) {
					if u.Username != "jane" || u.Email != "jane@example.com" {
						t.Errorf("patch not lowercased: %+v", u)
					}
					return u, tt.mockErr
				},
			)

			svc := NewService(repo)
			_, err := svc.Update(context.Background(), "1", "Jane", "Jane@Example.com")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr error
	}{
		{name: "deleted"},
		{name: "not found", mockErr: ErrNotFound, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().Delete(gomock.Any(), "1").Return(tt.mockErr)

			svc := NewService(repo)
			err := svc.Delete(context.Background(), "1")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
