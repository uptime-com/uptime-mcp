package cloudstatus

import (
	"testing"
)

func TestNewIndex(t *testing.T) {
	idx, err := NewIndex()
	if err != nil {
		t.Fatalf("NewIndex() error: %v", err)
	}

	providers := idx.Providers()
	if len(providers) == 0 {
		t.Fatal("expected at least one provider")
	}

	t.Logf("loaded %d providers", len(providers))

	// Verify well-known providers exist
	known := map[string]bool{"AWS": false, "Azure": false, "GCP": false}
	for _, p := range providers {
		if _, ok := known[p]; ok {
			known[p] = true
		}
	}
	for name, found := range known {
		if !found {
			t.Errorf("expected provider %q not found", name)
		}
	}
}

func TestSearch(t *testing.T) {
	idx, err := NewIndex()
	if err != nil {
		t.Fatalf("NewIndex() error: %v", err)
	}

	t.Run("by provider", func(t *testing.T) {
		result := idx.Search("AWS", "", 1, 10)
		if result.TotalCount == 0 {
			t.Fatal("expected AWS services")
		}
		for _, s := range result.Services {
			if s.Group != "AWS" {
				t.Errorf("expected group AWS, got %q", s.Group)
			}
		}
		t.Logf("AWS: %d total services", result.TotalCount)
	})

	t.Run("by search text", func(t *testing.T) {
		result := idx.Search("", "EC2", 1, 100)
		if result.TotalCount == 0 {
			t.Fatal("expected services matching EC2")
		}
		t.Logf("EC2 search: %d results", result.TotalCount)
	})

	t.Run("by provider and search", func(t *testing.T) {
		result := idx.Search("AWS", "EC2", 1, 100)
		if result.TotalCount == 0 {
			t.Fatal("expected AWS EC2 services")
		}
		for _, s := range result.Services {
			if s.Group != "AWS" {
				t.Errorf("expected group AWS, got %q", s.Group)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		full := idx.Search("AWS", "", 1, 10000)
		page1 := idx.Search("AWS", "", 1, 10)
		page2 := idx.Search("AWS", "", 2, 10)

		if page1.TotalCount != full.TotalCount {
			t.Errorf("total count mismatch: %d vs %d", page1.TotalCount, full.TotalCount)
		}
		if len(page1.Services) != 10 {
			t.Errorf("expected 10 results on page 1, got %d", len(page1.Services))
		}
		if len(page2.Services) == 0 {
			t.Error("expected results on page 2")
		}
		if page1.Services[0].Name == page2.Services[0].Name {
			t.Error("page 1 and 2 should return different results")
		}
	})

	t.Run("no match", func(t *testing.T) {
		result := idx.Search("NonExistentProvider", "", 1, 25)
		if result.TotalCount != 0 {
			t.Errorf("expected 0 results, got %d", result.TotalCount)
		}
	})
}
