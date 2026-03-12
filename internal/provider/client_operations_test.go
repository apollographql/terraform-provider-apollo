// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type expectedGraphQLExchange struct {
	operation string
	fixture   string
}

func TestClientDomainMethods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		exchanges []expectedGraphQLExchange
		run       func(context.Context, *Client) (any, error)
		assert    func(*testing.T, any)
	}{
		{
			name: "graph metadata",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetGraphMetadata, fixture: "graph_metadata_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.GetGraphMetadata(ctx, "inventory")
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				graph := value.(*GraphMetadata)
				if got, want := graph.ID, "inventory"; got != want {
					t.Fatalf("unexpected graph id: got %q want %q", got, want)
				}
				if got, want := graph.Name, "Inventory Graph"; got != want {
					t.Fatalf("unexpected graph name: got %q want %q", got, want)
				}
				if got, want := graph.Title, "Inventory"; got != want {
					t.Fatalf("unexpected graph title: got %q want %q", got, want)
				}
				if got, want := strings.Join(graph.VariantNames, ","), "current,staging"; got != want {
					t.Fatalf("unexpected variant list: got %q want %q", got, want)
				}
			},
		},
		{
			name: "variant metadata",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetVariantMetadata, fixture: "variant_metadata_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.GetVariantMetadata(ctx, "inventory", "current")
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				variant := value.(*VariantMetadata)
				if got, want := variant.ID, "variant-current"; got != want {
					t.Fatalf("unexpected variant id: got %q want %q", got, want)
				}
				if got, want := variant.Name, "current"; got != want {
					t.Fatalf("unexpected variant name: got %q want %q", got, want)
				}
				if got, want := len(variant.Subgraphs), 2; got != want {
					t.Fatalf("unexpected subgraph count: got %d want %d", got, want)
				}
				if got, want := variant.Subgraphs[0].Revision, "rev-products-1"; got != want {
					t.Fatalf("unexpected subgraph revision: got %q want %q", got, want)
				}
			},
		},
		{
			name: "delete graph variant",
			exchanges: []expectedGraphQLExchange{
				{operation: operationDeleteGraphVariant, fixture: "variant_delete_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return nil, client.DeleteGraphVariant(ctx, "inventory", "current")
			},
			assert: func(t *testing.T, _ any) {},
		},
		{
			name: "subgraph metadata",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetVariantMetadata, fixture: "variant_metadata_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.GetSubgraphMetadata(ctx, "inventory", "current", "products")
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				subgraph := value.(*SubgraphMetadata)
				if got, want := subgraph.Name, "products"; got != want {
					t.Fatalf("unexpected subgraph name: got %q want %q", got, want)
				}
				if got, want := subgraph.RoutingURL, "https://products.example.com/graphql"; got != want {
					t.Fatalf("unexpected routing url: got %q want %q", got, want)
				}
				if got, want := subgraph.Revision, "rev-products-1"; got != want {
					t.Fatalf("unexpected revision: got %q want %q", got, want)
				}
			},
		},
		{
			name: "publish subgraph schema",
			exchanges: []expectedGraphQLExchange{
				{operation: operationPublishSubgraphSchema, fixture: "publish_subgraph_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return nil, client.PublishSubgraphSchema(ctx, PublishSubgraphInput{
					GraphID:    "inventory",
					Variant:    "current",
					Subgraph:   "products",
					RoutingURL: "https://products.example.com/graphql",
					SchemaSDL:  "type Query { topProducts: [String!]! }",
					Revision:   "rev-products-1",
				})
			},
			assert: func(t *testing.T, _ any) {},
		},
		{
			name: "remove subgraph",
			exchanges: []expectedGraphQLExchange{
				{operation: operationRemoveSubgraph, fixture: "remove_subgraph_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return nil, client.RemoveSubgraph(ctx, "inventory", "current", "products")
			},
			assert: func(t *testing.T, _ any) {},
		},
		{
			name: "get persisted query list",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetPersistedQueryList, fixture: "pql_get_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.GetPersistedQueryList(ctx, "inventory", "pql-123")
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				pql := value.(*PersistedQueryList)
				if got, want := pql.Name, "mobile-clients"; got != want {
					t.Fatalf("unexpected pql name: got %q want %q", got, want)
				}
			},
		},
		{
			name: "get variant persisted query list",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetVariantPersistedQueryList, fixture: "pql_variant_get_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.GetVariantPersistedQueryList(ctx, "inventory", "current")
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				pql := value.(*PersistedQueryList)
				if got, want := pql.ID, "pql-123"; got != want {
					t.Fatalf("unexpected pql id: got %q want %q", got, want)
				}
				if !persistedQueryListHasVariantLink(pql, "inventory", "current") {
					t.Fatalf("expected variant link to be present in %#v", pql)
				}
			},
		},
		{
			name: "create persisted query list",
			exchanges: []expectedGraphQLExchange{
				{operation: operationCreatePersistedQueryList, fixture: "pql_create_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.CreatePersistedQueryList(ctx, PersistedQueryListInput{
					GraphID:     "inventory",
					Name:        "mobile-clients",
					Description: "Mobile clients",
				})
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				pql := value.(*PersistedQueryList)
				if got, want := pql.ID, "pql-123"; got != want {
					t.Fatalf("unexpected pql id: got %q want %q", got, want)
				}
			},
		},
		{
			name: "update persisted query list",
			exchanges: []expectedGraphQLExchange{
				{operation: operationUpdatePersistedQueryList, fixture: "pql_update_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.UpdatePersistedQueryList(ctx, PersistedQueryListInput{
					GraphID:     "inventory",
					ID:          "pql-123",
					Name:        "mobile-clients",
					Description: "Updated description",
				})
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				pql := value.(*PersistedQueryList)
				if got, want := pql.Name, "mobile-clients"; got != want {
					t.Fatalf("unexpected pql name: got %q want %q", got, want)
				}
			},
		},
		{
			name: "delete persisted query list",
			exchanges: []expectedGraphQLExchange{
				{operation: operationDeletePersistedQueryList, fixture: "pql_delete_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return nil, client.DeletePersistedQueryList(ctx, "inventory", "pql-123")
			},
			assert: func(t *testing.T, _ any) {},
		},
		{
			name: "link persisted query list to variant",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetVariantPersistedQueryList, fixture: "pql_variant_get_empty.json"},
				{operation: operationLinkPersistedQueryListVariant, fixture: "pql_link_success.json"},
				{operation: operationGetVariantPersistedQueryList, fixture: "pql_variant_get_success.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return client.LinkPersistedQueryListToVariant(ctx, PersistedQueryListLinkInput{
					PersistedQueryListID: "pql-123",
					GraphID:              "inventory",
					Variant:              "current",
				})
			},
			assert: func(t *testing.T, value any) {
				t.Helper()
				pql := value.(*PersistedQueryList)
				if got, want := pql.ID, "pql-123"; got != want {
					t.Fatalf("unexpected pql id: got %q want %q", got, want)
				}
			},
		},
		{
			name: "unlink persisted query list from variant",
			exchanges: []expectedGraphQLExchange{
				{operation: operationGetVariantPersistedQueryList, fixture: "pql_variant_get_success.json"},
				{operation: operationUnlinkPersistedQueryList, fixture: "pql_unlink_success.json"},
				{operation: operationGetVariantPersistedQueryList, fixture: "pql_variant_get_empty.json"},
			},
			run: func(ctx context.Context, client *Client) (any, error) {
				return nil, client.UnlinkPersistedQueryListFromVariant(ctx, PersistedQueryListLinkInput{
					PersistedQueryListID: "pql-123",
					GraphID:              "inventory",
					Variant:              "current",
				})
			},
			assert: func(t *testing.T, _ any) {},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			index := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if index >= len(testCase.exchanges) {
					t.Fatalf("received unexpected GraphQL request after %d expected exchanges", len(testCase.exchanges))
				}

				var request graphQLRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatalf("failed to decode GraphQL request: %s", err)
				}

				expected := testCase.exchanges[index]
				index++

				if got, want := request.OperationName, expected.operation; got != want {
					t.Fatalf("unexpected operation name at exchange %d: got %q want %q", index, got, want)
				}

				fixture, err := os.ReadFile(filepath.Join("testdata", expected.fixture))
				if err != nil {
					t.Fatalf("failed to load fixture: %s", err)
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(fixture)
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

			value, err := testCase.run(context.Background(), client)
			if err != nil {
				t.Fatalf("expected operation to succeed, got error: %s", err)
			}

			if index != len(testCase.exchanges) {
				t.Fatalf("expected %d exchanges, saw %d", len(testCase.exchanges), index)
			}

			testCase.assert(t, value)
		})
	}
}

func TestClientListPersistedQueryListsDeferred(t *testing.T) {
	t.Parallel()

	client := NewClient("service:key", "https://example.com/graphql", "terraform-provider-apollo/test")

	_, err := client.ListPersistedQueryLists(context.Background(), "inventory")
	if err == nil {
		t.Fatal("expected deferred capability error, got nil")
	}

	if !IsErrorKind(err, ErrorKindCapabilityDeferred) {
		t.Fatalf("expected deferred capability error, got %v", err)
	}
}

func TestClientDeleteGraphVariantNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationDeleteGraphVariant; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"graph":{"variant":null}}}`))
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

	err := client.DeleteGraphVariant(context.Background(), "inventory", "missing")
	if err == nil {
		t.Fatal("expected not found error")
	}

	if !IsErrorKind(err, ErrorKindNotFound) {
		t.Fatalf("expected not found error, got %T %v", err, err)
	}
}

func TestClientDeleteGraphVariantPermissionDenied(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationDeleteGraphVariant; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"permission denied","extensions":{"code":"PERMISSION_DENIED"}}]}`))
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

	err := client.DeleteGraphVariant(context.Background(), "inventory", "current")
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if !IsErrorKind(err, ErrorKindPermissionDenied) {
		t.Fatalf("expected permission denied error, got %T %v", err, err)
	}
}

func TestClientDeleteGraphVariantMalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationDeleteGraphVariant; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"graph":{"variant":{"delete":{"deleted":false}}}}}`))
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

	err := client.DeleteGraphVariant(context.Background(), "inventory", "current")
	if err == nil {
		t.Fatal("expected malformed response error")
	}

	if !IsErrorKind(err, ErrorKindMalformedResponse) {
		t.Fatalf("expected malformed response error, got %T %v", err, err)
	}
}

func TestClientRemoveSubgraphNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationRemoveSubgraph; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"subgraph not found","extensions":{"code":"NOT_FOUND","status":404}}]}`))
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

	err := client.RemoveSubgraph(context.Background(), "inventory", "current", "missing")
	if err == nil {
		t.Fatal("expected not found error")
	}

	if !IsErrorKind(err, ErrorKindNotFound) {
		t.Fatalf("expected not found error, got %T %v", err, err)
	}
}

func TestClientRemoveSubgraphPermissionDenied(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationRemoveSubgraph; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"permission denied","extensions":{"code":"PERMISSION_DENIED"}}]}`))
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

	err := client.RemoveSubgraph(context.Background(), "inventory", "current", "products")
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	if !IsErrorKind(err, ErrorKindPermissionDenied) {
		t.Fatalf("expected permission denied error, got %T %v", err, err)
	}
}

func TestClientRemoveSubgraphMalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationRemoveSubgraph; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"graph":{"removeImplementingServiceAndTriggerComposition":null}}}`))
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

	err := client.RemoveSubgraph(context.Background(), "inventory", "current", "products")
	if err == nil {
		t.Fatal("expected malformed response error")
	}

	if !IsErrorKind(err, ErrorKindMalformedResponse) {
		t.Fatalf("expected malformed response error, got %T %v", err, err)
	}
}

func TestClientRemoveSubgraphCompositionFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode GraphQL request: %s", err)
		}

		if got, want := request.OperationName, operationRemoveSubgraph; got != want {
			t.Fatalf("unexpected operation name: got %q want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"graph":{"removeImplementingServiceAndTriggerComposition":{"updatedGateway":false,"errors":[{"code":"SATISFIABILITY_ERROR","message":"cannot compose without products"}]}}}}`))
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

	err := client.RemoveSubgraph(context.Background(), "inventory", "current", "products")
	if err == nil {
		t.Fatal("expected invalid request error")
	}

	if !IsErrorKind(err, ErrorKindInvalidRequest) {
		t.Fatalf("expected invalid request error, got %T %v", err, err)
	}

	if !strings.Contains(err.Error(), "cannot compose without products") {
		t.Fatalf("expected composition failure detail, got %v", err)
	}
}

func TestClientRemoveSubgraphSucceedsWhenSubgraphIsAlreadyGone(t *testing.T) {
	t.Parallel()

	server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
		{
			operation: operationRemoveSubgraph,
			payload: map[string]any{
				"data": map[string]any{
					"graph": nil,
				},
				"errors": []map[string]any{
					{
						"message": "Unable to call pipelineorchestrator",
						"extensions": map[string]any{
							"code": "INTERNAL_SERVER_ERROR",
						},
					},
				},
			},
		},
		{
			operation: operationGetVariantMetadata,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"variant": map[string]any{
							"id":        "inventory@current",
							"name":      "current",
							"subgraphs": []map[string]any{},
						},
					},
				},
			},
		},
	})
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:      "service:key",
		Endpoint:    server.URL(),
		UserAgent:   "terraform-provider-apollo/test",
		HTTPClient:  server.server.Client(),
		MaxAttempts: 1,
		RetryDelay:  func(int) time.Duration { return 0 },
	})

	if err := client.RemoveSubgraph(context.Background(), "inventory", "current", "products"); err != nil {
		t.Fatalf("expected subgraph removal to be treated as success after post-delete read, got %s", err)
	}
}
