package validation

import (
    "net/http"
    "testing"
    "net/http/httptest"
)

func TestLoadSpec(t *testing.T) {
    spec, err := LoadSpec("../../api/openapi.yaml")
    if err != nil {
        t.Fatalf("LoadSpec() error = %v", err)
    }

    if spec == nil {
        t.Fatal("LoadSpec() returned nil spec")
    }
}

func TestInvalidMD5(t *testing.T) {
    spec, err := LoadSpec("../../api/openapi.yaml")
    if err != nil {
        t.Fatal(err)
    }

    middleware, err := NewMiddleware(spec)
    if err != nil {
        t.Fatal(err)
    }

    handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest(http.MethodGet, "/info/invalid-md5", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusBadRequest {
        t.Fatalf("expected status 400, got %d", rec.Code)
    }
}
