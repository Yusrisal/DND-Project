package config

import (
	"os"
	"testing"
)

func TestLoadDSN_IndividualVars(t *testing.T) {
	os.Unsetenv("DB_DSN")
	os.Setenv("DB_HOST", "192.168.1.15")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "appuser")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "dnd_db")
	os.Setenv("DB_CHARSET", "utf8mb4")
	os.Setenv("DB_PARSE_TIME", "true")
	os.Setenv("DB_LOC", "Local")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_CHARSET")
		os.Unsetenv("DB_PARSE_TIME")
		os.Unsetenv("DB_LOC")
	}()

	cfg := Load()
	want := "appuser:secret@tcp(192.168.1.15:3306)/dnd_db?charset=utf8mb4&parseTime=true&loc=Local"
	if cfg.DBDSN != want {
		t.Fatalf("got %q, want %q", cfg.DBDSN, want)
	}
}

func TestLoadDSN_EmptyPassword(t *testing.T) {
	os.Unsetenv("DB_DSN")
	os.Setenv("DB_HOST", "192.168.1.15")
	os.Setenv("DB_USER", "appuser")
	os.Unsetenv("DB_PASSWORD")
	os.Setenv("DB_NAME", "dnd_db")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_NAME")
	}()

	cfg := Load()
	want := "appuser@tcp(192.168.1.15:3306)/dnd_db?charset=utf8mb4&parseTime=true&loc=Local"
	if cfg.DBDSN != want {
		t.Fatalf("got %q, want %q", cfg.DBDSN, want)
	}
}

func TestLoadDSN_Fallback(t *testing.T) {
	os.Unsetenv("DB_DSN")
	os.Unsetenv("DB_HOST")
	cfg := Load()
	want := "dnd:dnd@tcp(127.0.0.1:3306)/dnd?parseTime=true&charset=utf8mb4&loc=Local"
	if cfg.DBDSN != want {
		t.Fatalf("got %q, want %q", cfg.DBDSN, want)
	}
}
