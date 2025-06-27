package test_support

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"dogecoin.org/fractal-engine/pkg/config"
	"dogecoin.org/fractal-engine/pkg/store"
)

func SetupTestDB(t *testing.T) *store.TokenisationStore {
	testDir := os.TempDir()
	dbPath := filepath.Join(testDir, fmt.Sprintf("test_rpc_%d.db", rand.Intn(1000000)))

	tokenisationStore, err := store.NewTokenisationStore("sqlite:///"+dbPath, config.Config{
		MigrationsPath: "../../../db/migrations",
	})
	if err != nil {
		t.Fatalf("Failed to create tokenisation store: %v", err)
	}

	err = tokenisationStore.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate tokenisation store: %v", err)
	}

	return tokenisationStore
}
