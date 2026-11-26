package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"threat-service/api"
	"threat-service/internal/handler"
	"threat-service/internal/repository"
	"threat-service/internal/service"
	"threat-service/internal/validation"
)

func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()

	repo := repository.NewInMemoryThreatRepository()
	svc := service.NewThreatService(repo)
	h := handler.NewThreatHandler(svc)

	router := api.Handler(h)

	spec, err := validation.LoadSpec("../../api/openapi.yaml")
	if len(spec.Servers) > 0 {
		spec.Servers[0].URL = "/"
	}
	if err != nil {
		t.Fatalf("failed to load OpenAPI spec: %v", err)
	}

	middleware, err := validation.NewMiddleware(spec)
	if err != nil {
		t.Fatalf("failed to create validation middleware: %v", err)
	}

	server := httptest.NewServer(middleware(router))
	t.Cleanup(server.Close)

	return server
}

func requestJSON(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	payload any,
) (int, []byte) {
	t.Helper()

	var body io.Reader

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return resp.StatusCode, respBody
}

func TestThreatAPIWorkflow(t *testing.T) {
	server := newIntegrationServer(t)
	client := server.Client()

	md5 := strings.Repeat("a", 32)
	url := fmt.Sprintf("%s/info/%s", server.URL, md5)

	createRequest := map[string]any{
		"md5":         md5,
		"name":        "Test Trojan",
		"description": "Initial description",
		"severity":    "low",
	}

	status, body := requestJSON(
		t, client, http.MethodPost,
		server.URL+"/info", createRequest,
	)

	if status != http.StatusCreated {
		t.Fatalf("POST: expected 201, got %d: %s", status, body)
	}

	status, body = requestJSON(
		t, client, http.MethodGet, url, nil,
	)

	if status != http.StatusOK {
		t.Fatalf("GET: expected 200, got %d", status)
	}

	var threat struct {
		Md5         string  `json:"md5"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Severity    *string `json:"severity"`
	}

	if err := json.Unmarshal(body, &threat); err != nil {
		t.Fatalf("failed to decode GET response: %v", err)
	}

	if threat.Md5 != md5 {
		t.Errorf("expected MD5 %s, got %s", md5, threat.Md5)
	}

	if threat.Name != "Test Trojan" {
		t.Errorf("expected name Test Trojan, got %s", threat.Name)
	}

	if threat.Description == nil ||
		*threat.Description != "Initial description" {
		t.Error("initial description was not preserved")
	}

	updateRequest := map[string]any{
		"name":     "Updated Trojan",
		"severity": "high",
	}

	status, _ = requestJSON(
		t, client, http.MethodPut, url, updateRequest,
	)

	if status != http.StatusNoContent {
		t.Fatalf("PUT: expected 204, got %d", status)
	}

	status, body = requestJSON(
		t, client, http.MethodGet, url, nil,
	)

	if status != http.StatusOK {
		t.Fatalf("GET after update: expected 200, got %d", status)
	}

	if err := json.Unmarshal(body, &threat); err != nil {
		t.Fatalf("failed to decode updated threat: %v", err)
	}

	if threat.Name != "Updated Trojan" {
		t.Errorf("expected updated name, got %s", threat.Name)
	}

	if threat.Severity == nil || *threat.Severity != "high" {
		t.Error("severity was not updated correctly")
	}

	if threat.Description == nil ||
		*threat.Description != "Initial description" {
		t.Error("update did not preserve omitted description")
	}

	status, _ = requestJSON(
		t, client, http.MethodPost,
		server.URL+"/info", createRequest,
	)

	if status != http.StatusConflict {
		t.Errorf("duplicate POST: expected 409, got %d", status)
	}

	missingMD5 := strings.Repeat("b", 32)
	missingURL := fmt.Sprintf("%s/info/%s", server.URL, missingMD5)

	status, _ = requestJSON(
		t, client, http.MethodGet, missingURL, nil,
	)

	if status != http.StatusNotFound {
		t.Errorf("missing GET: expected 404, got %d", status)
	}
}

func TestThreatAPIValidation(t *testing.T) {
	server := newIntegrationServer(t)
	client := server.Client()

	t.Run("InvalidMD5", func(t *testing.T) {
		status, _ := requestJSON(
			t,
			client,
			http.MethodGet,
			server.URL+"/info/invalid-md5",
			nil,
		)

		if status != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", status)
		}
	})

	t.Run("InvalidSeverity", func(t *testing.T) {
		payload := map[string]any{
			"md5":      strings.Repeat("a", 32),
			"name":     "Test Trojan",
			"severity": "extreme",
		}

		status, _ := requestJSON(
			t, client, http.MethodPost,
			server.URL+"/info", payload,
		)

		if status != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", status)
		}
	})

	t.Run("UnknownField", func(t *testing.T) {
		payload := map[string]any{
			"md5":          strings.Repeat("b", 32),
			"name":         "Test Trojan",
			"unknownField": "unexpected",
		}

		status, _ := requestJSON(
			t, client, http.MethodPost,
			server.URL+"/info", payload,
		)

		if status != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", status)
		}
	})
}
