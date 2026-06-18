package catalog

import (
	"context"
	"testing"

	"ziniao/internal/apperr"
)

func TestMockProviderListsModules(t *testing.T) {
	provider := NewMockProvider()

	items, err := provider.ListModules(context.Background())
	if err != nil {
		t.Fatalf("ListModules() error = %v", err)
	}
	if len(items) == 0 || items[0].Name != "ziniao" {
		t.Fatalf("modules = %#v", items)
	}
}

func TestMockProviderGetsAPIDocument(t *testing.T) {
	provider := NewMockProvider()

	doc, err := provider.GetAPI(context.Background(), "ziniao", "user", "list")
	if err != nil {
		t.Fatalf("GetAPI() error = %v", err)
	}
	if doc.URL != "/api/user/list" || doc.Method != "GET" {
		t.Fatalf("doc = %#v", doc)
	}
}

func TestMockProviderReportsUnknownAPI(t *testing.T) {
	provider := NewMockProvider()

	_, err := provider.GetAPI(context.Background(), "ziniao", "user", "missing")
	if err == nil {
		t.Fatal("GetAPI() error = nil, want error")
	}
	if got := apperr.Kind(err); got != apperr.KindAPI {
		t.Fatalf("error kind = %q", got)
	}
}
