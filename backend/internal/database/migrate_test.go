package database

import (
	"testing"
)

func TestCleanStatement(t *testing.T) {
	input := `
-- One encounter per map at a time; active is cleared when combat ends.
CREATE TABLE encounters (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (id)
)`

	want := `CREATE TABLE encounters (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (id)
)`

	got := cleanStatement(input)
	if got != want {
		t.Fatalf("cleanStatement mismatch.\nGot:\n%s\nWant:\n%s", got, want)
	}
}

func TestCleanStatement_OnlyComments(t *testing.T) {
	input := `
-- Tokens gain their own vitals
-- Another comment
`
	got := cleanStatement(input)
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
