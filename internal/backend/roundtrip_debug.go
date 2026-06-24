package backend

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"ziniao/internal/debug"
)

const debugBodyMaxLen = 8 << 10

type loggingRoundTripper struct {
	base http.RoundTripper
}

func (t *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBody, err := readAndRestoreBody(&req.Body)
	if err != nil {
		return nil, err
	}

	debug.Printf("[debug] --> %s %s\n", req.Method, req.URL.String())
	for key, values := range req.Header {
		for _, value := range values {
			debug.Printf("[debug]     %s: %s\n", key, debug.RedactHeader(key, value))
		}
	}
	if len(reqBody) > 0 {
		debug.Printf("[debug]     body: %s\n", formatDebugBody(reqBody))
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		debug.Printf("[debug] <-- error: %v\n", err)
		return resp, err
	}

	respBody, err := readAndRestoreBody(&resp.Body)
	if err != nil {
		return nil, err
	}

	debug.Printf("[debug] <-- %s\n", resp.Status)
	if len(respBody) > 0 {
		debug.Printf("[debug]     body: %s\n", formatDebugBody(respBody))
	}

	return resp, nil
}

func readAndRestoreBody(body *io.ReadCloser) ([]byte, error) {
	if body == nil || *body == nil {
		return nil, nil
	}

	payload, err := io.ReadAll(*body)
	if closeErr := (*body).Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}

	*body = io.NopCloser(bytes.NewReader(payload))
	return payload, nil
}

func formatDebugBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	display := body
	truncated := false
	if len(display) > debugBodyMaxLen {
		display = display[:debugBodyMaxLen]
		truncated = true
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, display, "", "  "); err == nil {
		if truncated {
			return pretty.String() + "\n... (truncated)"
		}
		return pretty.String()
	}

	text := string(display)
	if truncated {
		return text + "\n... (truncated)"
	}
	return text
}

func newHTTPTransport() http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	var roundTripper http.RoundTripper = transport
	if debug.Enabled() {
		roundTripper = &loggingRoundTripper{base: transport}
	}
	return roundTripper
}
