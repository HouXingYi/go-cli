package catalog

import (
	"encoding/json"
	"testing"
)

func TestParseModules(t *testing.T) {
	raw := json.RawMessage(`{"items":[{"name":"ziniao","title":"紫鸟业务接口"}]}`)
	items, err := ParseModules(raw)
	if err != nil {
		t.Fatalf("ParseModules() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "ziniao" {
		t.Fatalf("items = %#v", items)
	}
}

func TestParseAPIDocument(t *testing.T) {
	raw := json.RawMessage(`{"name":"list","title":"查询用户列表","url":"/api/user/list","method":"GET"}`)
	doc, err := ParseAPIDocument(raw)
	if err != nil {
		t.Fatalf("ParseAPIDocument() error = %v", err)
	}
	if doc.URL != "/api/user/list" || doc.Method != "GET" {
		t.Fatalf("doc = %#v", doc)
	}
}

func TestParseFullAPIsFromArray(t *testing.T) {
	raw := json.RawMessage(`[{"name":"list","method":"GET","url":"/api/user/list"}]`)
	items, err := ParseFullAPIs(raw)
	if err != nil {
		t.Fatalf("ParseFullAPIs() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "list" {
		t.Fatalf("items = %#v", items)
	}
}
