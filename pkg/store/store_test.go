package store

import (
	"database/sql"
	"testing"

	"dogecoin.org/fractal-engine/pkg/protocol"
	"gotest.tools/assert"
)

func TestStore(t *testing.T) {
	store, err := NewStore("memory://:memory:")
	if err != nil {
		t.Fatal(err)
	}

	err = store.Migrate()
	if err != nil {
		t.Fatal(err)
	}

	id, err := store.SaveMint(&protocol.MintWithoutID{
		Title:           "Test Mint",
		Description:     "Test Description",
		Tags:            []string{"test"},
		Metadata:        map[string]interface{}{"test": "test"},
		Hash:            "test",
		Verified:        false,
		OutputAddress:   "test",
		TransactionHash: sql.NullString{String: "test", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	mints, err := store.GetMints(0, 10, false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(mints))
	assert.Equal(t, id, mints[0].Id)
	assert.Equal(t, "Test Mint", mints[0].Title)
}
