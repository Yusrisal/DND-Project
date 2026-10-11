package database

import (
	"testing"
)

func TestStripComments(t *testing.T) {
	input := `-- The bestiary. A monster here is a template; spawning copies its vitals onto a
-- token so each goblin takes its own damage, while monster_id keeps the link
CREATE TABLE monsters (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (id)
);
-- Another comment with semicolon; test
ALTER TABLE tokens ADD COLUMN test INT;`

	want := `CREATE TABLE monsters (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (id)
);
ALTER TABLE tokens ADD COLUMN test INT;`

	got := stripComments(input)
	if got != want {
		t.Fatalf("stripComments mismatch.\nGot:\n%s\nWant:\n%s", got, want)
	}
}
