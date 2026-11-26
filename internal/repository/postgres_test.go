package repository

import (
	"context"
	"os"
	"testing"

	"threat-service/api"
)

func newTestPostgresRepository(t *testing.T) *PostgresThreatRepository {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; skipping PostgreSQL tests")
	}

	repo, err := NewPostgresThreatRepository(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	t.Cleanup(repo.Close)

	return repo
}

func TestPostgresCreateAndGet(t *testing.T) {
	repo := newTestPostgresRepository(t)

	md5 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	t.Cleanup(func() {
		_, err := repo.pool.Exec(
			context.Background(),
			"DELETE FROM threats WHERE md5 = $1",
			md5,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	})

	threat := api.Threat{
		Md5:  md5,
		Name: "Test Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.Get(md5)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Md5 != threat.Md5 {
		t.Errorf("expected MD5 %s, got %s", threat.Md5, got.Md5)
	}

	if got.Name != threat.Name {
		t.Errorf("expected name %s, got %s", threat.Name, got.Name)
	}
}

func TestPostgresCreateDuplicate(t *testing.T) {
	repo := newTestPostgresRepository(t)

	md5 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	t.Cleanup(func() {
		_, err := repo.pool.Exec(
			context.Background(),
			"DELETE FROM threats WHERE md5 = $1",
			md5,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	})

	threat := api.Threat{
		Md5:  md5,
		Name: "Test Trojan",
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	err := repo.Create(threat)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestPostgresUpdate(t *testing.T) {
	repo := newTestPostgresRepository(t)

	md5 := "cccccccccccccccccccccccccccccccc"

	t.Cleanup(func() {
		_, err := repo.pool.Exec(
			context.Background(),
			"DELETE FROM threats WHERE md5 = $1",
			md5,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	})

	original := api.Threat{
		Md5:  md5,
		Name: "Original Name",
	}

	if err := repo.Create(original); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	updated := api.Threat{
		Md5:  md5,
		Name: "Updated Name",
	}

	if err := repo.Update(md5, updated); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.Get(md5)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Name != updated.Name {
		t.Errorf("expected name %q, got %q", updated.Name, got.Name)
	}
}

func TestPostgresGetNotFound(t *testing.T) {
	repo := newTestPostgresRepository(t)

	_, err := repo.Get("dddddddddddddddddddddddddddddddd")

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresUpdateNotFound(t *testing.T) {
	repo := newTestPostgresRepository(t)

	threat := api.Threat{
		Md5:  "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		Name: "Nonexistent Threat",
	}

	err := repo.Update(threat.Md5, threat)

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresNullableFieldsAndTags(t *testing.T) {
	repo := newTestPostgresRepository(t)

	md5 := "ffffffffffffffffffffffffffffffff"

	t.Cleanup(func() {
		_, err := repo.pool.Exec(
			context.Background(),
			"DELETE FROM threats WHERE md5 = $1",
			md5,
		)
		if err != nil {
			t.Errorf("failed to clean up test data: %v", err)
		}
	})

	description := "Suspicious executable"
	tags := []string{"malware", "trojan"}

	threat := api.Threat{
		Md5:         md5,
		Name:        "Test Threat",
		Description: &description,
		Tags:        &tags,
	}

	if err := repo.Create(threat); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.Get(md5)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Description == nil || *got.Description != description {
		t.Errorf("unexpected description: %v", got.Description)
	}

	if got.Tags == nil || len(*got.Tags) != len(tags) {
		t.Fatalf("unexpected tags: %v", got.Tags)
	}

	for i, tag := range tags {
		if (*got.Tags)[i] != tag {
			t.Errorf("expected tag %q at index %d, got %q", tag, i, (*got.Tags)[i])
		}
	}

	if got.Severity != nil {
		t.Errorf("expected nil severity, got %v", *got.Severity)
	}
}
