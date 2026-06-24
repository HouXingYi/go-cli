package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ziniao/internal/config"
	"ziniao/internal/debug"
)

func TestNewHTTPTransportWithoutDebug(t *testing.T) {
	t.Setenv(debug.EnvDebug, "")

	rt := newHTTPTransport()
	if _, ok := rt.(*loggingRoundTripper); ok {
		t.Fatal("newHTTPTransport() wrapped logging round tripper when debug disabled")
	}
}

func TestNewHTTPTransportWithDebug(t *testing.T) {
	t.Setenv(debug.EnvDebug, "1")

	rt := newHTTPTransport()
	if _, ok := rt.(*loggingRoundTripper); !ok {
		t.Fatalf("newHTTPTransport() = %T, want *loggingRoundTripper", rt)
	}
}

func TestHTTPBackendDebugLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ret":0,"status":"success","msg":"ok","data":{"ok":true}}`))
	}))
	defer server.Close()

	var debugBuf bytes.Buffer
	oldOutput := debug.Output
	debug.Output = &debugBuf
	t.Cleanup(func() {
		debug.Output = oldOutput
	})

	t.Run("disabled", func(t *testing.T) {
		debugBuf.Reset()
		t.Setenv(debug.EnvDebug, "")

		be := NewHTTPBackend(config.Config{
			ProxyBaseURL: server.URL,
			AuthKey:      "test-key",
		})

		raw, err := be.Proxy(context.Background(), "ziniao", "GET", "/user/list", json.RawMessage(`{"page":1}`), nil)
		if err != nil {
			t.Fatalf("Proxy() error = %v", err)
		}
		if string(raw) != `{"ok":true}` {
			t.Fatalf("raw = %s", raw)
		}
		if debugBuf.Len() != 0 {
			t.Fatalf("debug output = %q, want empty", debugBuf.String())
		}
	})

	t.Run("enabled", func(t *testing.T) {
		debugBuf.Reset()
		t.Setenv(debug.EnvDebug, "1")

		be := NewHTTPBackend(config.Config{
			ProxyBaseURL: server.URL,
			AuthKey:      "secret-key",
		})

		raw, err := be.Proxy(context.Background(), "ziniao", "GET", "/user/list", json.RawMessage(`{"page":1}`), nil)
		if err != nil {
			t.Fatalf("Proxy() error = %v", err)
		}
		if string(raw) != `{"ok":true}` {
			t.Fatalf("raw = %s", raw)
		}

		logs := debugBuf.String()
		if !strings.Contains(logs, "[debug] --> POST "+server.URL+"/cli/ziniao/user/list") {
			t.Fatalf("missing request log: %q", logs)
		}
		if strings.Contains(logs, "secret-key") {
			t.Fatalf("auth key leaked in logs: %q", logs)
		}
		if !strings.Contains(logs, "Bearer ***") {
			t.Fatalf("missing redacted auth header: %q", logs)
		}
		if !strings.Contains(logs, `"method": "GET"`) {
			t.Fatalf("missing request body in logs: %q", logs)
		}
		if !strings.Contains(logs, "[debug] <-- 200 OK") {
			t.Fatalf("missing response status in logs: %q", logs)
		}
		if !strings.Contains(logs, `"ok": true`) {
			t.Fatalf("missing response body in logs: %q", logs)
		}
	})
}

func TestFormatDebugBodyTruncatesLargePayload(t *testing.T) {
	large := []byte(`{"value":"` + strings.Repeat("a", debugBodyMaxLen) + `"}`)
	got := formatDebugBody(large)
	if !strings.Contains(got, "... (truncated)") {
		t.Fatalf("formatDebugBody() = %q, want truncation marker", got)
	}
}

func TestFormatDebugBodyPrettyPrintsJSON(t *testing.T) {
	got := formatDebugBody([]byte(`{"ok":true,"items":[1,2]}`))
	if got != "{\n  \"ok\": true,\n  \"items\": [\n    1,\n    2\n  ]\n}" {
		t.Fatalf("formatDebugBody() = %q", got)
	}
}
