// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientDoGraphQLRequestShape(t *testing.T) {
	t.Parallel()

	var requestBody graphQLRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("x-api-key"), "service:key"; got != want {
			t.Fatalf("unexpected api key header: got %q want %q", got, want)
		}

		if got, want := r.Header.Get("User-Agent"), "terraform-provider-apollo/test"; got != want {
			t.Fatalf("unexpected user agent: got %q want %q", got, want)
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("failed to decode request body: %s", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.Client(),
		MaxAttempts: 1,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "TestOperation", "query TestOperation($id: ID!) { ok(id: $id) }", map[string]any{"id": "graph-1"}, &response)
	if err != nil {
		t.Fatalf("expected request to succeed, got error: %s", err)
	}

	if !response.OK {
		t.Fatal("expected response payload to decode")
	}

	if got, want := requestBody.OperationName, "TestOperation"; got != want {
		t.Fatalf("unexpected operation name: got %q want %q", got, want)
	}

	if got, want := requestBody.Variables["id"], "graph-1"; got != want {
		t.Fatalf("unexpected variable payload: got %#v want %#v", got, want)
	}
}

func TestClientDoGraphQLRetriesRateLimit(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAttempt := atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")

		if currentAttempt == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errors":[{"message":"too many requests","extensions":{"code":"RATE_LIMITED"}}]}`))
			return
		}

		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.Client(),
		MaxAttempts: 2,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "RetryRateLimit", "query RetryRateLimit { ok }", nil, &response)
	if err != nil {
		t.Fatalf("expected retry path to succeed, got error: %s", err)
	}

	if got, want := atomic.LoadInt32(&attempts), int32(2); got != want {
		t.Fatalf("unexpected attempt count: got %d want %d", got, want)
	}
}

func TestClientDoGraphQLRetriesServerError(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAttempt := atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")

		if currentAttempt == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"errors":[{"message":"upstream failed"}]}`))
			return
		}

		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.Client(),
		MaxAttempts: 2,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "RetryServerError", "query RetryServerError { ok }", nil, &response)
	if err != nil {
		t.Fatalf("expected retry path to succeed, got error: %s", err)
	}

	if got, want := atomic.LoadInt32(&attempts), int32(2); got != want {
		t.Fatalf("unexpected attempt count: got %d want %d", got, want)
	}
}

func TestClientDoGraphQLDoesNotRetryGraphQLValidationFailure(t *testing.T) {
	t.Parallel()

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"validation failed","extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}]}`))
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.Client(),
		MaxAttempts: 3,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "ValidationFailure", "query ValidationFailure { ok }", nil, &response)
	if err == nil {
		t.Fatal("expected GraphQL validation error")
	}

	if !IsErrorKind(err, ErrorKindInvalidRequest) {
		t.Fatalf("expected invalid request error, got %T %v", err, err)
	}

	if got, want := atomic.LoadInt32(&attempts), int32(1); got != want {
		t.Fatalf("unexpected attempt count: got %d want %d", got, want)
	}
}

func TestClientDoGraphQLTimeoutHandling(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	httpClient := server.Client()
	httpClient.Timeout = 10 * time.Millisecond

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  httpClient,
		MaxAttempts: 1,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "Timeout", "query Timeout { ok }", nil, &response)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !IsErrorKind(err, ErrorKindTransient) {
		t.Fatalf("expected transient timeout error, got %T %v", err, err)
	}
}

func TestClientDoGraphQLMalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL,
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.Client(),
		MaxAttempts: 1,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	var response struct {
		OK bool `json:"ok"`
	}
	err := client.doGraphQL(context.Background(), "Malformed", "query Malformed { ok }", nil, &response)
	if err == nil {
		t.Fatal("expected malformed response error")
	}

	if !IsErrorKind(err, ErrorKindMalformedResponse) {
		t.Fatalf("expected malformed response error, got %T %v", err, err)
	}
}
