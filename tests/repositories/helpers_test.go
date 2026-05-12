package repositories_test

import (
"context"
"os"
"testing"

"github.com/jackc/pgx/v5/pgxpool"
)

// newTestDB opens a connection pool for integration tests.
// Skips the test if TEST_DATABASE_URL is not set.
func newTestDB(t *testing.T) *pgxpool.Pool {
t.Helper()
url := os.Getenv("TEST_DATABASE_URL")
if url == "" {
t.Skip("TEST_DATABASE_URL not set — skipping integration test")
}
pool, err := pgxpool.New(context.Background(), url)
if err != nil {
t.Fatalf("connect to test DB: %v", err)
}
t.Cleanup(func() { pool.Close() })
return pool
}
