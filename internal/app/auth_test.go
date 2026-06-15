package app

import (
	"context"
	"testing"
)

type fakeClient struct {
	verifyCalled bool
	runCalled    bool
}

func (f *fakeClient) VerifyAuth(ctx context.Context) (AuthVerification, error) {
	f.verifyCalled = true
	return AuthVerification{Message: "verified"}, nil
}

func (f *fakeClient) RunRequest(ctx context.Context) (RequestResult, error) {
	f.runCalled = true
	return RequestResult{Message: "done"}, nil
}

func TestVerifyAuthRequiresToken(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client, "")

	_, err := service.VerifyAuth(context.Background())
	if err == nil {
		t.Fatal("VerifyAuth() error = nil, want error")
	}
	if client.verifyCalled {
		t.Fatal("VerifyAuth() called API without token")
	}
}

func TestRunRequestCallsAPI(t *testing.T) {
	client := &fakeClient{}
	service := NewService(client, "test-token")

	result, err := service.RunRequest(context.Background())
	if err != nil {
		t.Fatalf("RunRequest() error = %v", err)
	}
	if !client.runCalled {
		t.Fatal("RunRequest() did not call API")
	}
	if result.Message != "done" {
		t.Fatalf("Message = %q", result.Message)
	}
}
