package service

import (
	"errors"
	"testing"
	"threat-service/api"
	"threat-service/internal/repository"
)

func TestUpdatePreservesOmittedFields(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := NewThreatService(repo)

	description := "Original description"
	severity := api.ThreatSeverityHigh

	threat := api.Threat{
		Md5:         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name:        "Original Trojan",
		Description: &description,
		Severity:    &severity,
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	request := api.ThreatUpdateRequest{
		Name: "Updated Trojan",
	}

	if err := svc.Update(threat.Md5, request); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := svc.Get(threat.Md5)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.Name != "Updated Trojan" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Trojan")
	}

	if got.Description == nil || *got.Description != description {
		t.Errorf("Description was not preserved")
	}

	if got.Severity == nil || *got.Severity != severity {
		t.Errorf("Severity was not preserved")
	}
}

func TestUpdateNotFound(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := NewThreatService(repo)

	request := api.ThreatUpdateRequest{
		Name: "Updated Trojan",
	}

	err := svc.Update("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", request)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Update() error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestCreate(t *testing.T) {
	repo := repository.NewInMemoryThreatRepository()
	svc := NewThreatService(repo)

	description := "Test description"

	request := api.ThreatCreateRequest{
		Md5:         "cccccccccccccccccccccccccccccccc",
		Name:        "Test Trojan",
		Description: &description,
	}

	if err := svc.Create(request); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := svc.Get(request.Md5)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.Name != request.Name {
		t.Errorf("Name = %q, want %q", got.Name, request.Name)
	}

	if got.Description == nil || *got.Description != description {
		t.Errorf("Description was not preserved")
	}

	if got.CreatedAt == nil {
		t.Error("CreatedAt was not set")
	}

	if got.UpdatedAt == nil {
		t.Error("UpdatedAt was not set")
	}
}
