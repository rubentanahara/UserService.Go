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

func TestService_List(t *testing.T) {
	tests := []struct {
		name      string
		limit     int64
		offset    int64
		wantLimit int64
		mockErr   error
		wantErr   error
	}{
		{name: "passthrough within max", limit: 10, offset: 5, wantLimit: 10},
		{name: "clamps zero to max", limit: 0, wantLimit: maxListLimit},
		{name: "clamps negative to max", limit: -1, wantLimit: maxListLimit},
		{name: "clamps over-max to max", limit: maxListLimit + 1, wantLimit: maxListLimit},
		{name: "propagates repo error", limit: 10, wantLimit: 10, mockErr: errors.New("boom"), wantErr: errors.New("boom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), tt.wantLimit, tt.offset).Return([]*User{}, tt.mockErr)

			svc := NewService(repo)
			_, err := svc.List(context.Background(), tt.limit, tt.offset)
			if (err == nil) != (tt.wantErr == nil) {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Authenticate(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	stored := &User{Email: "jane@example.com", Password: string(hash)}

	tests := []struct {
		name        string
		password    string
		mockUser    *User
		mockErr     error
		wantErr     error
		wantWrapped bool
	}{
		{name: "correct credentials", password: "password123", mockUser: stored},
		{name: "wrong password", password: "wrong", mockUser: stored, wantErr: ErrInvalidCredentials},
		{name: "unknown email", password: "password123", mockErr: ErrNotFound, wantErr: ErrInvalidCredentials},
		{name: "repo error", password: "password123", mockErr: errors.New("boom"), wantWrapped: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().GetByEmail(gomock.Any(), "jane@example.com").Return(tt.mockUser, tt.mockErr)

			svc := NewService(repo)
			u, err := svc.Authenticate(context.Background(), "Jane@Example.com", tt.password)

			switch {
			case tt.wantWrapped:
				if err == nil || errors.Is(err, ErrInvalidCredentials) {
					t.Fatalf("err = %v, want a wrapped non-sentinel error", err)
				}
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
			default:
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				if u != stored {
					t.Fatalf("user = %v, want %v", u, stored)
				}
			}
		})
	}
}

func TestService_ChangePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	stored := &User{Password: string(hash)}

	tests := []struct {
		name        string
		getByIDUser *User
		getByIDErr  error
		oldPassword string
		wantErr     error
		wantUpdate  bool
	}{
		{name: "changed", getByIDUser: stored, oldPassword: "oldpassword", wantUpdate: true},
		{name: "wrong old password", getByIDUser: stored, oldPassword: "wrong", wantErr: ErrInvalidCredentials},
		{name: "user not found", getByIDErr: ErrNotFound, oldPassword: "oldpassword", wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewMockRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), "1").Return(tt.getByIDUser, tt.getByIDErr)
			if tt.wantUpdate {
				repo.EXPECT().UpdatePassword(gomock.Any(), "1", gomock.Any()).DoAndReturn(
					func(_ context.Context, _, newHash string) error {
						if err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte("newpassword123")); err != nil {
							t.Errorf("stored hash does not match new password: %v", err)
						}
						return nil
					},
				)
			}

			svc := NewService(repo)
			err := svc.ChangePassword(context.Background(), "1", tt.oldPassword, "newpassword123")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
