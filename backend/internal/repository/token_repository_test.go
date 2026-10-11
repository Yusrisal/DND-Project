package repository

import (
	"strings"
	"testing"
)

// scanToken passes a fixed number of destinations to Scan, so a column added to
// the SELECT without a matching field is a runtime error no compiler catches.
const tokenScanFields = 16

func TestTokenColumnsMatchScanArity(t *testing.T) {
	if got := topLevelFields(tokenColumns); got != tokenScanFields {
		t.Fatalf("tokenColumns selects %d columns, scanToken reads %d", got, tokenScanFields)
	}
}

// The join to characters makes a bare `id` or `name` ambiguous, which MySQL
// only complains about at query time.
func TestTokenColumnsAreQualified(t *testing.T) {
	for _, field := range strings.Split(flatten(tokenColumns), ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if !strings.HasPrefix(field, "t.") && !strings.HasPrefix(field, "COALESCE(") {
			t.Errorf("column %q is not table-qualified", field)
		}
	}
}

// topLevelFields counts commas outside parentheses, so COALESCE(a, b) is one.
func topLevelFields(columns string) int {
	depth, count := 0, 1
	for _, r := range columns {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}

// flatten replaces the contents of every parenthesised group so the splitter
// above can work on commas alone.
func flatten(columns string) string {
	var out strings.Builder
	depth := 0
	for _, r := range columns {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth > 0 && r == ',' {
			continue
		}
		if r == '\n' || r == '\t' {
			r = ' '
		}
		out.WriteRune(r)
	}
	return out.String()
}
