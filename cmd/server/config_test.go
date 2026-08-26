package main

import (
	"testing"
)

func TestGetenv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		fallback string
		want     string
	}{
		{name: "uses env when set", envValue: "custom", setEnv: true, fallback: "default", want: "custom"},
		{name: "uses fallback when unset", setEnv: false, fallback: "default", want: "default"},
		{name: "uses fallback when empty", envValue: "", setEnv: true, fallback: "default", want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("TEST_GETENV_KEY", tt.envValue)
			}
			if got := getenv("TEST_GETENV_KEY", tt.fallback); got != tt.want {
				t.Fatalf("getenv() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProvideMongoURI(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	if got := provideMongoURI(); got != "mongodb://localhost:27017" {
		t.Fatalf("default = %q, want %q", got, "mongodb://localhost:27017")
	}

	t.Setenv("MONGO_URI", "mongodb://custom:27017")
	if got := provideMongoURI(); got != "mongodb://custom:27017" {
		t.Fatalf("override = %q, want %q", got, "mongodb://custom:27017")
	}
}

func TestProvideMongoDBName(t *testing.T) {
	t.Setenv("MONGO_DB", "")
	if got := provideMongoDBName(); got != "user_service" {
		t.Fatalf("default = %q, want %q", got, "user_service")
	}

	t.Setenv("MONGO_DB", "custom_db")
	if got := provideMongoDBName(); got != "custom_db" {
		t.Fatalf("override = %q, want %q", got, "custom_db")
	}
}

func TestProvidePort(t *testing.T) {
	t.Setenv("PORT", "")
	if got := providePort(); got != "5001" {
		t.Fatalf("default = %q, want %q", got, "5001")
	}

	t.Setenv("PORT", "8080")
	if got := providePort(); got != "8080" {
		t.Fatalf("override = %q, want %q", got, "8080")
	}
}

func TestProvideMaxPoolSize(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    MaxPoolSize
		wantErr bool
	}{
		{name: "unset uses default", value: "", want: defaultMaxPoolSize},
		{name: "valid override", value: "200", want: 200},
		{name: "invalid value errors", value: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONGO_MAX_POOL_SIZE", tt.value)
			got, err := provideMaxPoolSize()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvideJWTSecret(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "unset errors", value: "", wantErr: true},
		{name: "set returns value", value: "my-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.value)
			got, err := provideJWTSecret()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && string(got) != tt.value {
				t.Fatalf("got = %q, want %q", got, tt.value)
			}
		})
	}
}
