package app

import (
	"context"
	"testing"

	"ziniao/internal/httpclient"
)

type fakeClient struct {
	requestCalled bool
}

func (f *fakeClient) Request(ctx context.Context, options httpclient.RequestOptions) (httpclient.Response, error) {
	f.requestCalled = true
	return httpclient.Response{StatusCode: 200, Status: "200 OK"}, nil
}

func TestHTTPRequestRequiresToken(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client, "")

	_, err := service.HTTPRequest(context.Background(), HTTPRequest{Method: "GET", Path: "/api/user/list"})
	if err == nil {
		t.Fatal("HTTPRequest() error = nil, want error")
	}
	if client.requestCalled {
		t.Fatal("HTTPRequest() called API without token")
	}
}

func TestHTTPRequestCallsAPI(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client, "test-token")

	result, err := service.HTTPRequest(context.Background(), HTTPRequest{Method: "GET", Path: "/api/user/list"})
	if err != nil {
		t.Fatalf("HTTPRequest() error = %v", err)
	}
	if !client.requestCalled {
		t.Fatal("HTTPRequest() did not call API")
	}
	if result.StatusCode != 200 {
		t.Fatalf("StatusCode = %d", result.StatusCode)
	}
}
