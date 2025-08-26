package rpc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithSecureAPI_AuthorizationScenarios(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
	}{
		{name: "no auth header", apiKey: "secret", authHeader: "", wantStatus: http.StatusForbidden, wantNextCalled: false},
		{name: "wrong scheme", apiKey: "secret", authHeader: "Token secret", wantStatus: http.StatusForbidden, wantNextCalled: false},
		{name: "bearer without space", apiKey: "secret", authHeader: "Bearer", wantStatus: http.StatusForbidden, wantNextCalled: false},
		{name: "wrong token", apiKey: "secret", authHeader: "Bearer nope", wantStatus: http.StatusForbidden, wantNextCalled: false},
		{name: "correct token", apiKey: "secret", authHeader: "Bearer secret", wantStatus: http.StatusOK, wantNextCalled: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			handler := withSecureAPI(tc.apiKey, next)

			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}

			if nextCalled != tc.wantNextCalled {
				t.Fatalf("nextCalled = %v, want %v", nextCalled, tc.wantNextCalled)
			}

			if tc.wantStatus == http.StatusOK {
				if rr.Body.String() != "ok" {
					t.Fatalf("body = %q, want %q", rr.Body.String(), "ok")
				}
			}
		})
	}
}
