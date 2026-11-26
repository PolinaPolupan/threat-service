package repository

import (
	"errors"
	"testing"
	"threat-service/api"
)

func TestCreateAndGet(t *testing.T) {
	repo := NewInMemoryThreatRepository()

	threat := api.Threat{
		Md5:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "Test Trojan",
	}

	err := repo.Create(threat)
	if err != nil {
		t.Fatalf("Create() returned an unexpected error: %v", err)
	}

	got, err := repo.Get(threat.Md5)
	if err != nil {
		t.Fatalf("Get() returned an unexpected error: %v", err)
	}

	if got.Name != threat.Name {
		t.Errorf("Get() returned name %q, want %q", got.Name, threat.Name)
	}
}

func TestCreateDuplicate(t *testing.T) {
	repo := NewInMemoryThreatRepository()

	threat := api.Threat{
		Md5:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "Test Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("first Create() returned an unexpected error: %v", err)
	}

	err := repo.Create(threat)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("second Create() error = %v, want %v", err, ErrAlreadyExists)
	}
}

func TestGetNotFound(t *testing.T) {
	repo := NewInMemoryThreatRepository()

	_, err := repo.Get("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, ErrNotFound)
	}
}

func TestUpdate(t *testing.T) {
	repo := NewInMemoryThreatRepository()

	threat := api.Threat{
		Md5:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "Test Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	updated := api.Threat{
		Md5:  threat.Md5,
		Name: "Updated Trojan",
	}

	if err := repo.Update(threat.Md5, updated); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := repo.Get(threat.Md5)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got.Name != "Updated Trojan" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Trojan")
	}
}

func TestUpdateNotFound(t *testing.T) {
	repo := NewInMemoryThreatRepository()

	threat := api.Threat{
		Md5:  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Name: "Unknown Threat",
	}

	err := repo.Update(threat.Md5, threat)

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Update() error = %v, want %v", err, ErrNotFound)
	}
}
