// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testServerModeCompositionFailure testServerMode = "composition_failure"

func TestAccSubgraphResourceDeferredCreate(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccSubgraphResourceConfig(server.URL()),
				ExpectError: regexp.MustCompile(`Create subgraph: capability deferred`),
			},
		},
	})
}

func TestAccSubgraphResourceImportReadDelete(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: " HTTPS://PRODUCTS.EXAMPLE.COM:443/ ",
		Revision:   " rev-products-1 ",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				Config: testAccSubgraphResourceConfig(server.URL()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_subgraph.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "variant", "current"),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "name", "products"),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "id", "inventory:current:products"),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "routing_url", "https://products.example.com"),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "revision", "rev-products-1"),
				),
			},
		},
	})

	if got := server.SubgraphCount(); got != 0 {
		t.Fatalf("expected imported subgraph to be deleted during destroy, got %d remaining", got)
	}
}

func TestAccSubgraphResourceRefreshMissingSubgraph(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				PreConfig:          func() { server.DeleteSubgraph("inventory", "current", "products") },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSubgraphResourceReadUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})
	server.SetMode(operationGetVariantMetadata, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccSubgraphResourceConfig(server.URL()),
				ResourceName:  "apollo_subgraph.test",
				ImportState:   true,
				ImportStateId: "inventory:current:products",
				ExpectError:   regexp.MustCompile(`Read subgraph: authentication failed`),
			},
		},
	})
}

func TestAccSubgraphResourceReadPermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})
	server.SetMode(operationGetVariantMetadata, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccSubgraphResourceConfig(server.URL()),
				ResourceName:  "apollo_subgraph.test",
				ImportState:   true,
				ImportStateId: "inventory:current:products",
				ExpectError:   regexp.MustCompile(`(?s)Read subgraph: permission denied.*Graph read or broader org access.*subgraph metadata`),
			},
		},
	})
}

func TestAccSubgraphResourceDeleteMissingSubgraph(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				Config:    testAccSubgraphResourceConfig(server.URL()),
				PreConfig: func() { server.DeleteSubgraph("inventory", "current", "products") },
				Destroy:   true,
			},
		},
	})
}

func TestAccSubgraphResourceDeleteUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				Config:      testAccSubgraphResourceConfig(server.URL()),
				PreConfig:   func() { server.SetMode(operationRemoveSubgraph, testServerModeUnauthenticated) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Delete subgraph: authentication failed`),
			},
		},
	})
}

func TestAccSubgraphResourceDeletePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				Config:      testAccSubgraphResourceConfig(server.URL()),
				PreConfig:   func() { server.SetMode(operationRemoveSubgraph, testServerModePermissionDenied) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)Delete subgraph: permission denied.*Graph admin or broader org access.*subgraph removal`),
			},
		},
	})
}

func TestAccSubgraphResourceDeleteCompositionFailure(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newSubgraphTestServer(t)
	defer server.Close()
	server.UpsertSubgraph(subgraphRecord{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphResourceConfig(server.URL()),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      "inventory:current:products",
				ImportStatePersist: true,
			},
			{
				Config:      testAccSubgraphResourceConfig(server.URL()),
				PreConfig:   func() { server.SetMode(operationRemoveSubgraph, testServerModeCompositionFailure) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)Delete subgraph.*produced composition.*errors`),
			},
		},
	})

	if got := server.SubgraphCount(); got != 0 {
		t.Fatalf("expected post-test cleanup to destroy the retained subgraph, got %d remaining", got)
	}
}

func TestSubgraphResourceModelFromAPI(t *testing.T) {
	t.Parallel()

	model := subgraphResourceModelFromAPI(&SubgraphMetadata{
		GraphID:    " inventory ",
		Variant:    " current ",
		Name:       " products ",
		RoutingURL: " HTTPS://PRODUCTS.EXAMPLE.COM:443/ ",
		Revision:   " rev-products-1 ",
	})

	if got, want := model.ID.ValueString(), "inventory:current:products"; got != want {
		t.Fatalf("unexpected id: got %q want %q", got, want)
	}

	if got, want := model.GraphID.ValueString(), "inventory"; got != want {
		t.Fatalf("unexpected graph id: got %q want %q", got, want)
	}

	if got, want := model.Variant.ValueString(), "current"; got != want {
		t.Fatalf("unexpected variant: got %q want %q", got, want)
	}

	if got, want := model.Name.ValueString(), "products"; got != want {
		t.Fatalf("unexpected name: got %q want %q", got, want)
	}

	if got, want := model.RoutingURL.ValueString(), "https://products.example.com"; got != want {
		t.Fatalf("unexpected routing url: got %q want %q", got, want)
	}

	if got, want := model.Revision.ValueString(), "rev-products-1"; got != want {
		t.Fatalf("unexpected revision: got %q want %q", got, want)
	}
}

func TestSubgraphResourceUpdateProviderBug(t *testing.T) {
	t.Parallel()

	resourceUnderTest := &subgraphResource{}
	var resp tfresource.UpdateResponse

	resourceUnderTest.Update(context.Background(), tfresource.UpdateRequest{}, &resp)

	if !diagnosticsContainSummary(resp.Diagnostics, "Update subgraph: provider bug") {
		t.Fatalf("expected provider bug diagnostic, got %#v", resp.Diagnostics)
	}
}

func TestSubgraphImportID(t *testing.T) {
	t.Parallel()

	parts, err := ParseImportID(" inventory : current : products ", 3)
	if err != nil {
		t.Fatalf("expected import ID to parse, got error: %s", err)
	}

	if got, want := JoinImportID(parts...), "inventory:current:products"; got != want {
		t.Fatalf("unexpected import id normalization: got %q want %q", got, want)
	}
}

func testAccSubgraphResourceConfig(endpoint string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

resource "apollo_subgraph" "test" {
  graph_id = "inventory"
  variant  = "current"
  name     = "products"
}
`, endpoint)
}

type subgraphRecord struct {
	GraphID    string
	Variant    string
	Name       string
	RoutingURL string
	Revision   string
}

type subgraphTestServer struct {
	t      *testing.T
	server *httptest.Server

	mu        sync.Mutex
	subgraphs map[string]subgraphRecord
	modes     map[string]testServerMode
}

func newSubgraphTestServer(t *testing.T) *subgraphTestServer {
	t.Helper()

	testServer := &subgraphTestServer{
		t:         t,
		subgraphs: map[string]subgraphRecord{},
		modes:     map[string]testServerMode{},
	}

	testServer.server = httptest.NewServer(http.HandlerFunc(testServer.handle))
	return testServer
}

func (s *subgraphTestServer) Close() {
	s.server.Close()
}

func (s *subgraphTestServer) URL() string {
	return s.server.URL
}

func (s *subgraphTestServer) SetMode(operation string, mode testServerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modes[operation] = mode
}

func (s *subgraphTestServer) UpsertSubgraph(record subgraphRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subgraphs[subgraphKey(record.GraphID, record.Variant, record.Name)] = record
}

func (s *subgraphTestServer) DeleteSubgraph(graphID string, variant string, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subgraphs, subgraphKey(graphID, variant, name))
}

func (s *subgraphTestServer) SubgraphCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subgraphs)
}

func (s *subgraphTestServer) modeForOperation(operation string) testServerMode {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := s.modes[operation]
	if mode == testServerModeUnauthenticated || mode == testServerModePermissionDenied {
		delete(s.modes, operation)
	}

	return mode
}

func (s *subgraphTestServer) handle(w http.ResponseWriter, r *http.Request) {
	s.t.Helper()

	var request graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.t.Fatalf("failed to decode GraphQL request: %s", err)
	}

	switch s.modeForOperation(request.OperationName) {
	case testServerModeUnauthenticated:
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(s.t, w, map[string]any{
			"errors": []map[string]any{
				{
					"message": "unauthorized",
					"extensions": map[string]any{
						"code": "UNAUTHENTICATED",
					},
				},
			},
		})
		return
	case testServerModePermissionDenied:
		w.WriteHeader(http.StatusForbidden)
		writeJSONResponse(s.t, w, map[string]any{
			"errors": []map[string]any{
				{
					"message": "permission denied",
					"extensions": map[string]any{
						"code": "PERMISSION_DENIED",
					},
				},
			},
		})
		return
	}

	switch request.OperationName {
	case operationGetVariantMetadata:
		s.handleGetVariantMetadata(w, request)
	case operationRemoveSubgraph:
		s.handleRemoveSubgraph(w, request)
	default:
		s.t.Fatalf("unexpected operation name in subgraph test server: %s", request.OperationName)
	}
}

func (s *subgraphTestServer) handleGetVariantMetadata(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variantName := variableString(request.Variables, "variantName")

	subgraphs := s.variantSubgraphs(graphID, variantName)
	if len(subgraphs) == 0 {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"id":      graphID,
					"variant": nil,
				},
			},
		})
		return
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"id": graphID,
				"variant": map[string]any{
					"id":        fmt.Sprintf("%s@%s", graphID, variantName),
					"name":      variantName,
					"subgraphs": subgraphs,
				},
			},
		},
	})
}

func (s *subgraphTestServer) handleRemoveSubgraph(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variantName := variableString(request.Variables, "variant")
	subgraphName := variableString(request.Variables, "subgraph")

	if s.modeForOperation(operationRemoveSubgraph) == testServerModeCompositionFailure {
		s.mu.Lock()
		delete(s.modes, operationRemoveSubgraph)
		s.mu.Unlock()

		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"removeImplementingServiceAndTriggerComposition": map[string]any{
						"updatedGateway": false,
						"errors": []map[string]any{
							{
								"code":    "SATISFIABILITY_ERROR",
								"message": "cannot compose without products",
							},
						},
					},
				},
			},
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := subgraphKey(graphID, variantName, subgraphName)
	if _, ok := s.subgraphs[key]; !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"errors": []map[string]any{
				{
					"message": "subgraph not found",
					"extensions": map[string]any{
						"code":   "NOT_FOUND",
						"status": 404,
					},
				},
			},
		})
		return
	}

	delete(s.subgraphs, key)
	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"removeImplementingServiceAndTriggerComposition": map[string]any{
					"updatedGateway": true,
					"errors":         []map[string]any{},
				},
			},
		},
	})
}

func (s *subgraphTestServer) variantSubgraphs(graphID string, variant string) []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	values := make([]map[string]any, 0)
	for _, record := range s.subgraphs {
		if record.GraphID != graphID || record.Variant != variant {
			continue
		}

		values = append(values, map[string]any{
			"name":     record.Name,
			"url":      record.RoutingURL,
			"revision": record.Revision,
		})
	}

	return values
}

func subgraphKey(graphID string, variant string, name string) string {
	return JoinImportID(graphID, variant, name)
}
