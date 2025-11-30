package sqlparser

import (
	"testing"
)

func TestCTEs(t *testing.T) {
	tests := []struct {
		name  string
		sql   string
		valid bool
	}{
		{
			name:  "simple CTE",
			sql:   "WITH cte AS (SELECT id FROM t) SELECT * FROM cte",
			valid: true,
		},
		{
			name:  "CTE with column list",
			sql:   "WITH cte (col1, col2) AS (SELECT id, name FROM t) SELECT * FROM cte",
			valid: true,
		},
		{
			name:  "multiple CTEs",
			sql:   "WITH cte1 AS (SELECT id FROM t1), cte2 AS (SELECT id FROM t2) SELECT * FROM cte1 JOIN cte2",
			valid: true,
		},
		{
			name:  "recursive CTE",
			sql:   "WITH RECURSIVE cte AS (SELECT 1 AS n UNION ALL SELECT n + 1 FROM cte WHERE n < 10) SELECT * FROM cte",
			valid: true,
		},
		{
			name:  "CTE with window function",
			sql:   "WITH sales AS (SELECT product, amount FROM orders) SELECT product, SUM(amount) OVER (ORDER BY product) FROM sales",
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
