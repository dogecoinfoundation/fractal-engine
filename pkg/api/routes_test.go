package api

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"dogecoin.org/fractal-engine/pkg/store"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

var routes *Routes
var dbStore *store.Store

func TestMain(t *testing.M) {
	router := gin.Default()
	localStore, err := store.NewStore("memory://file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}

	dbStore = localStore

	localStore.Migrate()

	routes = NewRoutes(router, dbStore)

	os.Exit(t.Run())
}

func TestCreateMint(t *testing.T) {
	withTransaction(dbStore, func() {
		ctx, w := GetTestGinContext("/mints")

		body := `{"title": "Test Mint", "description": "Test Description", "tags": ["test"], "metadata": {"test": "test"}}`
		ctx.Request.Body = io.NopCloser(strings.NewReader(body))

		routes.Route_CreateMint(ctx)

		assert.Equal(t, http.StatusOK, w.Code)

		mints, err := dbStore.GetMints(1, 0, false)

		if err != nil {
			t.Fatalf("Failed to get mints: %v", err)
		}

		assert.Equal(t, 1, len(mints))
		assert.Equal(t, "Test Mint", mints[0].Title)
	})
}

func TestCreateMintFailure(t *testing.T) {
	withTransaction(dbStore, func() {
		ctx, w := GetTestGinContext("/mints")

		routes.Route_CreateMint(ctx)
		ctx.Request.Body = io.NopCloser(strings.NewReader(""))

		mints, err := dbStore.GetMints(1, 0, false)
		if err != nil {
			t.Fatalf("Failed to get mints: %v", err)
		}

		assert.Equal(t, 0, len(mints))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func GetTestGinContext(path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Request = &http.Request{
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Method: "POST",
	}

	url, err := url.Parse(path)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}
	ctx.Request.URL = url

	return ctx, w
}

func withTransaction(store *store.Store, testFunc func()) {
	testFunc()

	clearTables(store.DB)
}

func clearTables(db *sql.DB) error {
	_, err := db.Exec(`
        DELETE FROM mints;
    `)
	return err
}
