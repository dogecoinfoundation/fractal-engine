package store_test

import (
	"os"
	"testing"

	"dogecoin.org/fractal-engine/internal/test/support"
)

// TestMain sets up and tears down the shared PostgreSQL container
func TestMain(m *testing.M) {
	// Initialize the shared PostgreSQL container
	if err := support.InitSharedPostgres(); err != nil {
		panic(err)
	}

	// Run all tests
	code := m.Run()

	// Cleanup is handled automatically by reference counting
	// The last test to finish will terminate the container

	os.Exit(code)
}

// Example test using the shared database
func TestExampleWithSharedDB(t *testing.T) {
	// This will create a unique database in the shared PostgreSQL instance
	store := support.SetupTestDB()

	// The database will be automatically dropped when the test finishes
	// thanks to t.Cleanup()

	// Your test logic here
	if store == nil {
		t.Fatal("Expected store to be initialized")
	}
}

// Another example test - each gets its own database
func TestAnotherExampleWithSharedDB(t *testing.T) {
	store := support.SetupTestDB()

	// This test has its own isolated database
	// No conflicts with other tests

	// Your test logic here
	if store == nil {
		t.Fatal("Expected store to be initialized")
	}
}
