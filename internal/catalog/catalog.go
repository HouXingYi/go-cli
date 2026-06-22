package catalog

import (
	"encoding/json"
	"fmt"

	"ziniao/internal/apperr"
)

type ModuleSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type BusinessSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type APISummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Method      string `json:"method"`
	URL         string `json:"url"`
}

type APIDocument struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	URL         string      `json:"url"`
	Method      string      `json:"method"`
	Params      interface{} `json:"params"`
	Response    interface{} `json:"response"`
}

func ParseModules(data json.RawMessage) ([]ModuleSummary, error) {
	var payload struct {
		Items []ModuleSummary `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, decodeError(err)
	}
	return payload.Items, nil
}

func ParseBusinesses(data json.RawMessage) ([]BusinessSummary, error) {
	var payload struct {
		Items []BusinessSummary `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, decodeError(err)
	}
	return payload.Items, nil
}

func ParseAPISummaries(data json.RawMessage) ([]APISummary, error) {
	var payload struct {
		Items []APISummary `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, decodeError(err)
	}
	return payload.Items, nil
}

func ParseFullAPIs(data json.RawMessage) ([]APIDocument, error) {
	var items []APIDocument
	if err := json.Unmarshal(data, &items); err == nil {
		return items, nil
	}
	var doc APIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, decodeError(err)
	}
	return []APIDocument{doc}, nil
}

func ParseAPIDocument(data json.RawMessage) (APIDocument, error) {
	var doc APIDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return APIDocument{}, decodeError(err)
	}
	return doc, nil
}

func decodeError(err error) error {
	return apperr.Wrap(apperr.KindAPI, "failed to parse catalog response", "check whether the API catalog response is valid.", err)
}

func FormatModules(items []ModuleSummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %s", item.Name, item.Title))
	}
	return joinLines(lines)
}

func FormatBusinesses(items []BusinessSummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %s", item.Name, item.Title))
	}
	return joinLines(lines)
}

func FormatAPISummaries(items []APISummary) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%-10s %-6s %-20s %s", item.Name, item.Method, item.URL, item.Title))
	}
	return joinLines(lines)
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
}
