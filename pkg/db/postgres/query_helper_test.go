package postgres

import (
	"testing"
)

func TestParseNamedQuery(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		args          map[string]any
		expectedQuery string
		expectedArgs  int
	}{
		{
			name:  "simple query with one parameter",
			query: "SELECT * FROM users WHERE id = :id",
			args: map[string]any{
				":id": 123,
			},
			expectedQuery: "SELECT * FROM users WHERE id =  $1 ",
			expectedArgs:  1,
		},
		{
			name:  "query with multiple parameters",
			query: "SELECT * FROM users WHERE id = :id AND name = :name",
			args: map[string]any{
				":id":   123,
				":name": "John",
			},
			expectedQuery: "SELECT * FROM users WHERE id =  $1  AND name =  $2 ",
			expectedArgs:  2,
		},
		{
			name:  "query with same parameter used twice",
			query: "SELECT * FROM users WHERE id = :id OR parent_id = :id",
			args: map[string]any{
				":id": 123,
			},
			expectedQuery: "SELECT * FROM users WHERE id =  $1  OR parent_id =  $1 ",
			expectedArgs:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.query
			params := ParseNamedQuery(&query, tt.args)

			if query != tt.expectedQuery {
				t.Errorf("Query mismatch.\nExpected: %q\nGot: %q", tt.expectedQuery, query)
			}

			if len(params) != tt.expectedArgs {
				t.Errorf("Expected %d params, got %d", tt.expectedArgs, len(params))
			}
		})
	}
}
