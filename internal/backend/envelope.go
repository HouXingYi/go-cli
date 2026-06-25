package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ziniao/internal/apperr"
)

type envelope struct {
	Ret    int             `json:"ret"`
	Status string          `json:"status"`
	Msg    string          `json:"msg"`
	Detail string          `json:"detail"`
	Data   json.RawMessage `json:"data"`
}

func parseEnvelope(statusCode int, body []byte) (json.RawMessage, error) {
	if statusCode == http.StatusUnauthorized {
		message := authErrorMessage(body)
		return nil, apperr.New(apperr.KindAuth, message, "check whether CLI_PROXY_TOKEN is valid.")
	}
	if statusCode < 200 || statusCode > 299 {
		return nil, apperr.New(apperr.KindAPI, httpErrorMessage(statusCode, body), "check the API response and request parameters.")
	}

	var payload envelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, apperr.Wrap(apperr.KindAPI, "failed to parse response", "check whether the API returns valid JSON.", err)
	}
	if payload.Ret != 0 {
		message := strings.TrimSpace(payload.Msg)
		if message == "" {
			message = fmt.Sprintf("api returned ret %d", payload.Ret)
		}
		return nil, apperr.New(apperr.KindAPI, message, retHint(payload.Ret))
	}
	if len(payload.Data) == 0 {
		return json.RawMessage("null"), nil
	}
	return payload.Data, nil
}

func authErrorMessage(body []byte) string {
	var payload struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Detail != "" {
			return payload.Detail
		}
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Msg != "" {
			return payload.Msg
		}
	}
	return "authentication failed"
}

func httpErrorMessage(statusCode int, body []byte) string {
	var payload struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Detail != "" {
			return payload.Detail
		}
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Msg != "" {
			return payload.Msg
		}
	}
	if len(body) == 0 {
		return fmt.Sprintf("api returned HTTP %d", statusCode)
	}
	return fmt.Sprintf("api returned HTTP %d: %s", statusCode, strings.TrimSpace(string(body)))
}

func retHint(ret int) string {
	switch ret {
	case 10000:
		return "upstream gateway authentication failed."
	case 30001:
		return "session is not bound to a user."
	case 30002:
		return "user has no auth credentials configured."
	case 30003:
		return "failed to obtain web session."
	case 30007:
		return "gateway request failed."
	default:
		return "check the API response and request parameters."
	}
}

func trimPath(path string) string {
	return strings.Trim(strings.TrimSpace(path), "/")
}

func rawMessageToMap(raw json.RawMessage) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
