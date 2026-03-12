// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGraphAPIKeyClientLifecycle(t *testing.T) {
	t.Parallel()

	server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
		{
			operation: operationCreateGraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if !strings.Contains(request.Query, "token") {
					t.Fatal("expected create graph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"newKey": map[string]any{
							"id":        "graph-key-1",
							"keyName":   "inventory-ci",
							"role":      "GRAPH_ADMIN",
							"createdAt": "2026-03-11T00:00:00Z",
							"lastUsed":  nil,
							"token":     "service:graph-token-1",
						},
					},
				},
			},
		},
		{
			operation: operationGetGraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if strings.Contains(request.Query, "token") {
					t.Fatal("did not expect read graph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"apiKeys": []map[string]any{
							{
								"id":        "graph-key-1",
								"keyName":   "inventory-ci",
								"role":      "GRAPH_ADMIN",
								"createdAt": "2026-03-11T00:00:00Z",
								"lastUsed":  "2026-03-12T00:00:00Z",
							},
						},
					},
				},
			},
		},
		{
			operation: operationUpdateGraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if strings.Contains(request.Query, "token") {
					t.Fatal("did not expect update graph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"renameKey": map[string]any{
							"id":        "graph-key-1",
							"keyName":   "inventory-deploy",
							"role":      "GRAPH_ADMIN",
							"createdAt": "2026-03-11T00:00:00Z",
							"lastUsed":  "2026-03-12T00:00:00Z",
						},
					},
				},
			},
		},
		{
			operation: operationDeleteGraphAPIKey,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"removeKey": nil,
					},
				},
			},
		},
	})
	defer server.Close()

	client := NewClientWithConfig(ClientConfig{
		APIKey:   "service:key",
		Endpoint: server.URL(),
	})

	key, err := client.CreateGraphAPIKey(context.Background(), GraphAPIKeyInput{
		GraphID: "inventory",
		Name:    "inventory-ci",
		Role:    "GRAPH_ADMIN",
	})
	if err != nil {
		t.Fatalf("expected create graph API key to succeed, got %s", err)
	}

	if got, want := key.Token, "service:graph-token-1"; got != want {
		t.Fatalf("unexpected token: got %q want %q", got, want)
	}

	key, err = client.GetGraphAPIKey(context.Background(), "inventory", "graph-key-1")
	if err != nil {
		t.Fatalf("expected get graph API key to succeed, got %s", err)
	}

	if got, want := key.LastUsedAt, "2026-03-12T00:00:00Z"; got != want {
		t.Fatalf("unexpected last used: got %q want %q", got, want)
	}

	key, err = client.UpdateGraphAPIKey(context.Background(), GraphAPIKeyInput{
		GraphID: "inventory",
		ID:      "graph-key-1",
		Name:    "inventory-deploy",
	})
	if err != nil {
		t.Fatalf("expected update graph API key to succeed, got %s", err)
	}

	if got, want := key.Name, "inventory-deploy"; got != want {
		t.Fatalf("unexpected updated name: got %q want %q", got, want)
	}

	if err := client.DeleteGraphAPIKey(context.Background(), "inventory", "graph-key-1"); err != nil {
		t.Fatalf("expected delete graph API key to succeed, got %s", err)
	}
}

func TestGraphAPIKeyClientErrors(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationGetGraphAPIKey,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id":      "inventory",
							"apiKeys": []map[string]any{},
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetGraphAPIKey(context.Background(), "inventory", "missing")
		if !IsErrorKind(err, ErrorKindNotFound) {
			t.Fatalf("expected not found error, got %#v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationCreateGraphAPIKey,
				status:    http.StatusUnauthorized,
				payload:   graphQLErrorPayload("unauthorized", "UNAUTHENTICATED", http.StatusUnauthorized),
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.CreateGraphAPIKey(context.Background(), GraphAPIKeyInput{
			GraphID: "inventory",
			Name:    "inventory-ci",
			Role:    "GRAPH_ADMIN",
		})
		if !IsErrorKind(err, ErrorKindUnauthenticated) {
			t.Fatalf("expected unauthenticated error, got %#v", err)
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationGetGraphAPIKey,
				status:    http.StatusForbidden,
				payload:   graphQLErrorPayload("permission denied", "PERMISSION_DENIED", http.StatusForbidden),
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetGraphAPIKey(context.Background(), "inventory", "graph-key-1")
		if !IsErrorKind(err, ErrorKindPermissionDenied) {
			t.Fatalf("expected permission denied error, got %#v", err)
		}
	})

	t.Run("not allowed by user role", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id": "inventory",
							"account": map[string]any{
								"id": "org-1",
							},
						},
					},
				},
			},
			{
				operation: operationGetSubgraphAPIKey,
				payload: map[string]any{
					"errors": []map[string]any{
						{
							"message": "NOT_ALLOWED_BY_USER_ROLE",
							"extensions": map[string]any{
								"code": "NOT_ALLOWED_BY_USER_ROLE",
							},
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			ID:           "subgraph-key-1",
		})
		if !IsErrorKind(err, ErrorKindPermissionDenied) {
			t.Fatalf("expected permission denied error, got %#v", err)
		}
	})

	t.Run("malformed create response", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationCreateGraphAPIKey,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"newKey": map[string]any{
								"id":        "graph-key-1",
								"keyName":   "inventory-ci",
								"role":      "GRAPH_ADMIN",
								"createdAt": "2026-03-11T00:00:00Z",
							},
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.CreateGraphAPIKey(context.Background(), GraphAPIKeyInput{
			GraphID: "inventory",
			Name:    "inventory-ci",
			Role:    "GRAPH_ADMIN",
		})
		if !IsErrorKind(err, ErrorKindMalformedResponse) {
			t.Fatalf("expected malformed response error, got %#v", err)
		}
	})
}

func TestSubgraphAPIKeyClientLifecycle(t *testing.T) {
	t.Parallel()

	server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
		{
			operation: operationResolveGraphOrgID,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"account": map[string]any{
							"id": "org-1",
						},
					},
				},
			},
		},
		{
			operation: operationCreateSubgraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if !strings.Contains(request.Query, "token") {
					t.Fatal("expected create subgraph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"organization": map[string]any{
						"createKey": map[string]any{
							"id":        "subgraph-key-1",
							"keyName":   "products-router",
							"keyType":   "SUBGRAPH",
							"createdAt": "2026-03-11T00:00:00Z",
							"lastUsed":  nil,
							"token":     "service:subgraph-token-1",
							"resources": []map[string]any{
								{
									"resourceId":   "inventory:production:products",
									"resourceType": "SUBGRAPH",
								},
							},
						},
					},
				},
			},
		},
		{
			operation: operationResolveGraphOrgID,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"account": map[string]any{
							"id": "org-1",
						},
					},
				},
			},
		},
		{
			operation: operationGetSubgraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if strings.Contains(request.Query, "token") {
					t.Fatal("did not expect read subgraph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"organization": map[string]any{
						"apiKey": map[string]any{
							"id":        "subgraph-key-1",
							"keyName":   "products-router",
							"keyType":   "SUBGRAPH",
							"createdAt": "2026-03-11T00:00:00Z",
							"lastUsed":  "2026-03-12T00:00:00Z",
							"resources": []map[string]any{
								{
									"resourceId":   "inventory:production:products",
									"resourceType": "SUBGRAPH",
								},
							},
						},
					},
				},
			},
		},
		{
			operation: operationResolveGraphOrgID,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"account": map[string]any{
							"id": "org-1",
						},
					},
				},
			},
		},
		{
			operation: operationUpdateSubgraphAPIKey,
			assert: func(t *testing.T, request graphQLRequest) {
				t.Helper()
				if strings.Contains(request.Query, "token") {
					t.Fatal("did not expect update subgraph API key query to request token")
				}
			},
			payload: map[string]any{
				"data": map[string]any{
					"organization": map[string]any{
						"renameKey": map[string]any{
							"id":        "subgraph-key-1",
							"keyName":   "products-runtime",
							"keyType":   "SUBGRAPH",
							"createdAt": "2026-03-11T00:00:00Z",
							"lastUsed":  "2026-03-12T00:00:00Z",
							"resources": []map[string]any{
								{
									"resourceId":   "inventory:production:products",
									"resourceType": "SUBGRAPH",
								},
							},
						},
					},
				},
			},
		},
		{
			operation: operationResolveGraphOrgID,
			payload: map[string]any{
				"data": map[string]any{
					"graph": map[string]any{
						"id": "inventory",
						"account": map[string]any{
							"id": "org-1",
						},
					},
				},
			},
		},
		{
			operation: operationDeleteSubgraphAPIKey,
			payload: map[string]any{
				"data": map[string]any{
					"organization": map[string]any{
						"deleteKey": "subgraph-key-1",
					},
				},
			},
		},
	})
	defer server.Close()

	client := NewClient("service:key", server.URL(), "")
	key, err := client.CreateSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
		GraphID:      "inventory",
		Variant:      "production",
		SubgraphName: "products",
		Name:         "products-router",
	})
	if err != nil {
		t.Fatalf("expected create subgraph API key to succeed, got %s", err)
	}

	if got, want := key.Token, "service:subgraph-token-1"; got != want {
		t.Fatalf("unexpected token: got %q want %q", got, want)
	}

	key, err = client.GetSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
		GraphID:      "inventory",
		Variant:      "production",
		SubgraphName: "products",
		ID:           "subgraph-key-1",
	})
	if err != nil {
		t.Fatalf("expected get subgraph API key to succeed, got %s", err)
	}

	if got, want := key.LastUsedAt, "2026-03-12T00:00:00Z"; got != want {
		t.Fatalf("unexpected last used: got %q want %q", got, want)
	}

	key, err = client.UpdateSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
		GraphID:      "inventory",
		Variant:      "production",
		SubgraphName: "products",
		ID:           "subgraph-key-1",
		Name:         "products-runtime",
	})
	if err != nil {
		t.Fatalf("expected update subgraph API key to succeed, got %s", err)
	}

	if got, want := key.Name, "products-runtime"; got != want {
		t.Fatalf("unexpected updated name: got %q want %q", got, want)
	}

	if err := client.DeleteSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
		GraphID: "inventory",
		ID:      "subgraph-key-1",
	}); err != nil {
		t.Fatalf("expected delete subgraph API key to succeed, got %s", err)
	}
}

func TestParseSubgraphAPIKeyResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		resourceID string
		graphID    string
		variant    string
		subgraph   string
		wantErr    bool
	}{
		{
			name:       "provider contract format",
			resourceID: "inventory:production:products",
			graphID:    "inventory",
			variant:    "production",
			subgraph:   "products",
		},
		{
			name:       "live GraphOS namespaced format",
			resourceID: "po.mdg:mattr-terraform:current:products",
			graphID:    "mattr-terraform",
			variant:    "current",
			subgraph:   "products",
		},
		{
			name:       "invalid format",
			resourceID: "inventory:products",
			wantErr:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			graphID, variant, subgraph, err := parseSubgraphAPIKeyResourceID(testCase.resourceID)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected parse error for %q", testCase.resourceID)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected parse to succeed, got %s", err)
			}

			if graphID != testCase.graphID || variant != testCase.variant || subgraph != testCase.subgraph {
				t.Fatalf("unexpected parse result: got %q/%q/%q want %q/%q/%q", graphID, variant, subgraph, testCase.graphID, testCase.variant, testCase.subgraph)
			}
		})
	}
}

func TestSubgraphAPIKeyClientErrors(t *testing.T) {
	t.Parallel()

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id": "inventory",
							"account": map[string]any{
								"id": "org-1",
							},
						},
					},
				},
			},
			{
				operation: operationGetSubgraphAPIKey,
				payload: map[string]any{
					"data": map[string]any{
						"organization": map[string]any{
							"apiKey": nil,
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			ID:           "missing",
		})
		if !IsErrorKind(err, ErrorKindNotFound) {
			t.Fatalf("expected not found error, got %#v", err)
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id": "inventory",
							"account": map[string]any{
								"id": "org-1",
							},
						},
					},
				},
			},
			{
				operation: operationGetSubgraphAPIKey,
				status:    http.StatusForbidden,
				payload:   graphQLErrorPayload("permission denied", "PERMISSION_DENIED", http.StatusForbidden),
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			ID:           "subgraph-key-1",
		})
		if !IsErrorKind(err, ErrorKindPermissionDenied) {
			t.Fatalf("expected permission denied error, got %#v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				status:    http.StatusUnauthorized,
				payload:   graphQLErrorPayload("unauthorized", "UNAUTHENTICATED", http.StatusUnauthorized),
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.CreateSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			Name:         "products-router",
		})
		if !IsErrorKind(err, ErrorKindUnauthenticated) {
			t.Fatalf("expected unauthenticated error, got %#v", err)
		}
	})

	t.Run("multi target unsupported", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id": "inventory",
							"account": map[string]any{
								"id": "org-1",
							},
						},
					},
				},
			},
			{
				operation: operationGetSubgraphAPIKey,
				payload: map[string]any{
					"data": map[string]any{
						"organization": map[string]any{
							"apiKey": map[string]any{
								"id":        "subgraph-key-1",
								"keyName":   "products-router",
								"keyType":   "SUBGRAPH",
								"createdAt": "2026-03-11T00:00:00Z",
								"lastUsed":  nil,
								"resources": []map[string]any{
									{
										"resourceId":   "inventory:production:products",
										"resourceType": "SUBGRAPH",
									},
									{
										"resourceId":   "inventory:staging:products",
										"resourceType": "SUBGRAPH",
									},
								},
							},
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.GetSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			ID:           "subgraph-key-1",
		})
		if !IsErrorKind(err, ErrorKindCapabilityDeferred) {
			t.Fatalf("expected capability deferred error, got %#v", err)
		}
	})

	t.Run("malformed create response", func(t *testing.T) {
		t.Parallel()

		server := newGraphQLSequenceServer(t, []graphQLSequenceStep{
			{
				operation: operationResolveGraphOrgID,
				payload: map[string]any{
					"data": map[string]any{
						"graph": map[string]any{
							"id": "inventory",
							"account": map[string]any{
								"id": "org-1",
							},
						},
					},
				},
			},
			{
				operation: operationCreateSubgraphAPIKey,
				payload: map[string]any{
					"data": map[string]any{
						"organization": map[string]any{
							"createKey": map[string]any{
								"id":        "subgraph-key-1",
								"keyName":   "products-router",
								"keyType":   "SUBGRAPH",
								"createdAt": "2026-03-11T00:00:00Z",
								"resources": []map[string]any{
									{
										"resourceId":   "inventory:production:products",
										"resourceType": "SUBGRAPH",
									},
								},
							},
						},
					},
				},
			},
		})
		defer server.Close()

		client := NewClient("service:key", server.URL(), "")
		_, err := client.CreateSubgraphAPIKey(context.Background(), SubgraphAPIKeyInput{
			GraphID:      "inventory",
			Variant:      "production",
			SubgraphName: "products",
			Name:         "products-router",
		})
		if !IsErrorKind(err, ErrorKindMalformedResponse) {
			t.Fatalf("expected malformed response error, got %#v", err)
		}
	})
}

type graphQLSequenceStep struct {
	operation string
	assert    func(*testing.T, graphQLRequest)
	status    int
	payload   map[string]any
}

type graphQLSequenceServer struct {
	t       *testing.T
	server  *httptest.Server
	steps   []graphQLSequenceStep
	current int
}

func newGraphQLSequenceServer(t *testing.T, steps []graphQLSequenceStep) *graphQLSequenceServer {
	t.Helper()

	server := &graphQLSequenceServer{
		t:     t,
		steps: append([]graphQLSequenceStep(nil), steps...),
	}

	server.server = httptest.NewServer(http.HandlerFunc(server.handle))
	return server
}

func (s *graphQLSequenceServer) Close() {
	s.server.Close()
	if s.current != len(s.steps) {
		s.t.Fatalf("expected %d GraphQL exchanges, saw %d", len(s.steps), s.current)
	}
}

func (s *graphQLSequenceServer) handle(w http.ResponseWriter, r *http.Request) {
	s.t.Helper()

	if s.current >= len(s.steps) {
		s.t.Fatalf("received unexpected GraphQL request after %d steps", s.current)
	}

	var request graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.t.Fatalf("failed to decode GraphQL request: %s", err)
	}

	step := s.steps[s.current]
	s.current++

	if request.OperationName != step.operation {
		s.t.Fatalf("unexpected operation at step %d: got %q want %q", s.current, request.OperationName, step.operation)
	}

	if step.assert != nil {
		step.assert(s.t, request)
	}

	if step.status != 0 {
		w.WriteHeader(step.status)
	}

	writeJSONResponse(s.t, w, step.payload)
}

func (s *graphQLSequenceServer) URL() string {
	return s.server.URL
}

func graphQLErrorPayload(message string, code string, status int) map[string]any {
	return map[string]any{
		"errors": []map[string]any{
			{
				"message": message,
				"extensions": map[string]any{
					"code":   code,
					"status": status,
				},
			},
		},
	}
}
