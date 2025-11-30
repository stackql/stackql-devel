package sqlparser

import (
	"testing"
)

func TestWindowFunctions(t *testing.T) {
	tests := []struct {
		name  string
		sql   string
		valid bool
	}{
		{
			name:  "simple window function",
			sql:   "SELECT SUM(count) OVER () FROM t",
			valid: true,
		},
		{
			name:  "window function with ORDER BY",
			sql:   "SELECT RANK() OVER (ORDER BY count DESC) FROM t",
			valid: true,
		},
		{
			name:  "window function with PARTITION BY",
			sql:   "SELECT SUM(count) OVER (PARTITION BY category) FROM t",
			valid: true,
		},
		{
			name:  "window function with PARTITION BY and ORDER BY",
			sql:   "SELECT SUM(count) OVER (PARTITION BY category ORDER BY name) FROM t",
			valid: true,
		},
		{
			name:  "window function with frame",
			sql:   "SELECT SUM(count) OVER (ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM t",
			valid: true,
		},
		{
			name:  "complex window function query",
			sql:   "SELECT serviceName, COUNT(*) as service_count, SUM(COUNT(*)) OVER () as total_count FROM t GROUP BY serviceName",
			valid: true,
		},
		{
			name:  "ROW_NUMBER window function",
			sql:   "SELECT ROW_NUMBER() OVER (ORDER BY id) as rn FROM t",
			valid: true,
		},
		{
			name:  "multiple window functions",
			sql:   "SELECT SUM(x) OVER (), COUNT(*) OVER (ORDER BY y) FROM t",
			valid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.sql)
			if tc.valid && err != nil {
				t.Errorf("expected valid SQL but got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("expected invalid SQL but got success")
			}
		})
	}
}
