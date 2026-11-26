package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"threat-service/api"
	"threat-service/internal/repository"
	"threat-service/internal/service"
)

func TestGetInfoByMd5(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	threat := api.Threat{
		Md5:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "Test Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/info/"+threat.Md5,
		nil,
	)

	rec := httptest.NewRecorder()

	h.GetInfoByMd5(rec, req, threat.Md5)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	if !strings.Contains(rec.Body.String(), `"name":"Test Trojan"`) {
		t.Errorf("response body does not contain expected name: %s", rec.Body.String())
	}
}

func TestCreateThreat(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	body := `{
		"md5": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"name": "Test Trojan"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/info",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.CreateThreat(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s",
			rec.Code, http.StatusCreated, rec.Body.String())
	}

	got, err := svc.Get("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("created threat not found: %v", err)
	}

	if got.Name != "Test Trojan" {
		t.Errorf("Name = %q, want %q", got.Name, "Test Trojan")
	}
}

func TestUpdateThreat(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	threat := api.Threat{
		Md5:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "Original Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	body := `{
		"name": "Updated Trojan"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/info/"+threat.Md5,
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.UpdateThreat(rec, req, threat.Md5)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body: %s",
			rec.Code, http.StatusNoContent, rec.Body.String())
	}

	got, err := svc.Get(threat.Md5)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.Name != "Updated Trojan" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Trojan")
	}
}

func TestGetInfoByMd5NotFound(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/info/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		nil,
	)

	rec := httptest.NewRecorder()

	h.GetInfoByMd5(rec, req, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCreateThreatDuplicate(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	body := `{
		"md5": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"name": "Test Trojan"
	}`

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/info",
			strings.NewReader(body),
		)

		rec := httptest.NewRecorder()
		h.CreateThreat(rec, req)

		expected := http.StatusCreated
		if i == 1 {
			expected = http.StatusConflict
		}

		if rec.Code != expected {
			t.Fatalf("request %d: status = %d, want %d",
				i+1, rec.Code, expected)
		}
	}
}

func TestCreateThreatInvalidJSON(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/info",
		strings.NewReader(`{"md5":`),
	)

	rec := httptest.NewRecorder()

	h.CreateThreat(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d",
			rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateThreatInvalidJSON(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	req := httptest.NewRequest(
		http.MethodPut,
		"/info/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		strings.NewReader(`{"name":`),
	)

	rec := httptest.NewRecorder()

	h.UpdateThreat(
		rec,
		req,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d",
			rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateThreatNotFound(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := NewThreatHandler(svc)

	body := `{"name":"Updated Trojan"}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/info/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.UpdateThreat(
		rec,
		req,
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d",
			rec.Code, http.StatusNotFound)
	}
}
