package main

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

func TestMigrationFilenamePattern(t *testing.T) {
	pattern := regexp.MustCompile(`^(\d{4})_(.+)\.sql$`)

	tests := []struct {
		filename string
		valid    bool
		version  int
		name     string
	}{
		{"0001_init_schema_migrations.sql", true, 1, "init_schema_migrations"},
		{"0042_add_indexes.sql", true, 42, "add_indexes"},
		{"001_invalid.sql", false, 0, ""},
		{"0001_test", false, 0, ""},
		{"0001.sql", false, 0, ""},
		{"invalid_0001_test.sql", false, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			matches := pattern.FindStringSubmatch(tt.filename)
			matched := matches != nil

			if matched != tt.valid {
				t.Errorf("filename %q: expected valid=%v, got valid=%v", tt.filename, tt.valid, matched)
				return
			}

			if tt.valid {
				version, err := strconv.Atoi(matches[1])
				if err != nil {
					t.Fatalf("parse version from %q: %v", matches[1], err)
				}
				if version != tt.version {
					t.Errorf("expected version %d, got %d", tt.version, version)
				}
				if matches[2] != tt.name {
					t.Errorf("expected name %q, got %q", tt.name, matches[2])
				}
			}
		})
	}
}

func TestMigrationChecksumConsistency(t *testing.T) {
	content := []byte("CREATE TABLE test (id INT64);")
	same := []byte("CREATE TABLE test (id INT64);")
	different := []byte("CREATE TABLE other (id INT64);")

	sum1 := fmt.Sprintf("%x", sha256.Sum256(content))
	sum2 := fmt.Sprintf("%x", sha256.Sum256(same))
	sum3 := fmt.Sprintf("%x", sha256.Sum256(different))

	if sum1 != sum2 {
		t.Error("same content must produce same checksum")
	}
	if sum1 == sum3 {
		t.Error("different content must not produce same checksum")
	}
}

func TestMigrationChecksumFormat(t *testing.T) {
	content := []byte("SELECT 1")
	checksum := fmt.Sprintf("%x", sha256.Sum256(content))

	// sha256 hex is always 64 chars
	if len(checksum) != 64 {
		t.Errorf("expected 64-char hex checksum, got %d chars: %s", len(checksum), checksum)
	}
}
