package model

import (
	"net/http/httptest"
	"testing"
)

// The pickers ask for the whole bestiary in one request. When the cap sat at
// 100 and the seeded bestiary grew to 105, the last five bosses silently
// vanished from the dropdowns with no error anywhere.
func TestParseQueryFitsTheSeededBestiary(t *testing.T) {
	const seededMonsters = 105

	r := httptest.NewRequest("GET", "/api/monsters?page_size=200", nil)
	q := ParseQuery(r, []string{"name"}, "name")

	if q.PageSize < seededMonsters {
		t.Fatalf("page size %d cannot hold the %d seeded monsters", q.PageSize, seededMonsters)
	}
}

func TestParseQueryClampsAndDefaults(t *testing.T) {
	cases := []struct {
		query string
		want  int
	}{
		{"", defaultPageSize},
		{"?page_size=50", 50},
		{"?page_size=99999", maxPageSize},
		{"?page_size=0", defaultPageSize},
		{"?page_size=-5", defaultPageSize},
		{"?page_size=abc", defaultPageSize},
	}

	for _, c := range cases {
		r := httptest.NewRequest("GET", "/api/things"+c.query, nil)
		if got := ParseQuery(r, []string{"name"}, "name").PageSize; got != c.want {
			t.Errorf("ParseQuery(%q).PageSize = %d, want %d", c.query, got, c.want)
		}
	}
}

// sort_by is interpolated into SQL, so anything off the allowlist must fall
// back to the default rather than reach the query.
func TestParseQueryRejectsSortColumnsOffTheAllowlist(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/things?sort_by=name;DROP+TABLE+users", nil)
	if got := ParseQuery(r, []string{"name", "created_at"}, "name").SortBy; got != "name" {
		t.Fatalf("SortBy = %q, want the default %q", got, "name")
	}
}
