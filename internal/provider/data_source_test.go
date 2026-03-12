// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphDataSourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphDataSourceConfig(server.URL(), "inventory"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollo_graph.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "name", "Inventory Graph"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "title", "Inventory"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "variant_names.#", "2"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "variant_names.0", "current"),
					resource.TestCheckResourceAttr("data.apollo_graph.test", "variant_names.1", "staging"),
				),
			},
		},
	})
}

func TestAccGraphDataSourceNotFound(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphDataSourceConfig(server.URL(), "missing"),
				ExpectError: regexp.MustCompile(`Read graph: remote object not found`),
			},
		},
	})
}

func TestAccGraphDataSourceUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetGraphMetadata, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphDataSourceConfig(server.URL(), "inventory"),
				ExpectError: regexp.MustCompile(`Read graph: authentication failed`),
			},
		},
	})
}

func TestAccGraphDataSourcePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetGraphMetadata, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphDataSourceConfig(server.URL(), "inventory"),
				ExpectError: regexp.MustCompile(`(?s)Read graph: permission denied.*Graph read or broader org access`),
			},
		},
	})
}

func TestAccGraphVariantDataSourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphVariantDataSourceConfig(server.URL(), "inventory", "current"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "variant", "current"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "id", "variant-current"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "subgraphs.#", "2"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "subgraphs.0.name", "products"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "subgraphs.0.routing_url", "https://products.example.com/graphql"),
					resource.TestCheckResourceAttr("data.apollo_graph_variant.test", "subgraphs.0.revision", "rev-products-1"),
				),
			},
		},
	})
}

func TestAccGraphVariantDataSourceNotFound(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphVariantDataSourceConfig(server.URL(), "inventory", "missing"),
				ExpectError: regexp.MustCompile(`Read graph variant: remote object not found`),
			},
		},
	})
}

func TestAccGraphVariantDataSourceUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetVariantMetadata, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphVariantDataSourceConfig(server.URL(), "inventory", "current"),
				ExpectError: regexp.MustCompile(`Read graph variant: authentication failed`),
			},
		},
	})
}

func TestAccGraphVariantDataSourcePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetVariantMetadata, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphVariantDataSourceConfig(server.URL(), "inventory", "current"),
				ExpectError: regexp.MustCompile(`(?s)Read graph variant: permission denied.*Graph read or broader org access.*variant metadata`),
			},
		},
	})
}

func TestAccSubgraphDataSourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSubgraphDataSourceConfig(server.URL(), "inventory", "current", "products"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "variant", "current"),
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "name", "products"),
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "id", "inventory:current:products"),
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "routing_url", "https://products.example.com/graphql"),
					resource.TestCheckResourceAttr("data.apollo_subgraph.test", "revision", "rev-products-1"),
				),
			},
		},
	})
}

func TestAccSubgraphDataSourceNotFound(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccSubgraphDataSourceConfig(server.URL(), "inventory", "current", "missing"),
				ExpectError: regexp.MustCompile(`Read subgraph: remote object not found`),
			},
		},
	})
}

func TestAccSubgraphDataSourceUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetVariantMetadata, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccSubgraphDataSourceConfig(server.URL(), "inventory", "current", "products"),
				ExpectError: regexp.MustCompile(`Read subgraph: authentication failed`),
			},
		},
	})
}

func TestAccSubgraphDataSourcePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetVariantMetadata, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccSubgraphDataSourceConfig(server.URL(), "inventory", "current", "products"),
				ExpectError: regexp.MustCompile(`(?s)Read subgraph: permission denied.*Graph read or broader org access.*subgraph metadata`),
			},
		},
	})
}

func TestAccPersistedQueryListDataSourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPersistedQueryListDataSourceConfig(server.URL(), "inventory", "pql-123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "id", "pql-123"),
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "name", "mobile-clients"),
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "linked_variants.#", "2"),
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "linked_variants.0.graph_id", "inventory"),
					resource.TestCheckResourceAttr("data.apollo_persisted_query_list.test", "linked_variants.0.variant", "current"),
				),
			},
		},
	})
}

func TestAccPersistedQueryListDataSourceNotFound(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccPersistedQueryListDataSourceConfig(server.URL(), "inventory", "missing"),
				ExpectError: regexp.MustCompile(`Read persisted query list: remote object not found`),
			},
		},
	})
}

func TestAccPersistedQueryListDataSourceUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetPersistedQueryList, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccPersistedQueryListDataSourceConfig(server.URL(), "inventory", "pql-123"),
				ExpectError: regexp.MustCompile(`Read persisted query list: authentication failed`),
			},
		},
	})
}

func TestAccPersistedQueryListDataSourcePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newMetadataDataSourceTestServer(t)
	defer server.Close()
	server.SetMode(operationGetPersistedQueryList, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccPersistedQueryListDataSourceConfig(server.URL(), "inventory", "pql-123"),
				ExpectError: regexp.MustCompile(`(?s)Read persisted query list: permission denied.*Graph or org access.*persisted query.*list management`),
			},
		},
	})
}

func testAccGraphDataSourceConfig(endpoint string, graphID string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

data "apollo_graph" "test" {
  graph_id = %q
}
`, endpoint, graphID)
}

func testAccGraphVariantDataSourceConfig(endpoint string, graphID string, variant string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

data "apollo_graph_variant" "test" {
  graph_id = %q
  variant  = %q
}
`, endpoint, graphID, variant)
}

func testAccSubgraphDataSourceConfig(endpoint string, graphID string, variant string, name string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

data "apollo_subgraph" "test" {
  graph_id = %q
  variant  = %q
  name     = %q
}
`, endpoint, graphID, variant, name)
}

func testAccPersistedQueryListDataSourceConfig(endpoint string, graphID string, id string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

data "apollo_persisted_query_list" "test" {
  graph_id = %q
  id       = %q
}
`, endpoint, graphID, id)
}

type metadataDataSourceTestServer struct {
	t      *testing.T
	server *httptest.Server

	mu    sync.Mutex
	modes map[string]testServerMode
}

func newMetadataDataSourceTestServer(t *testing.T) *metadataDataSourceTestServer {
	t.Helper()

	testServer := &metadataDataSourceTestServer{
		t:     t,
		modes: map[string]testServerMode{},
	}

	testServer.server = httptest.NewServer(http.HandlerFunc(testServer.handle))
	return testServer
}

func (s *metadataDataSourceTestServer) Close() {
	s.server.Close()
}

func (s *metadataDataSourceTestServer) URL() string {
	return s.server.URL
}

func (s *metadataDataSourceTestServer) SetMode(operation string, mode testServerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modes[operation] = mode
}

func (s *metadataDataSourceTestServer) modeForOperation(operation string) testServerMode {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := s.modes[operation]
	if mode != testServerModeNormal {
		delete(s.modes, operation)
	}

	return mode
}

func (s *metadataDataSourceTestServer) handle(w http.ResponseWriter, r *http.Request) {
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
	case operationGetGraphMetadata:
		s.handleGetGraphMetadata(w, request)
	case operationGetVariantMetadata:
		s.handleGetVariantMetadata(w, request)
	case operationGetPersistedQueryList:
		s.handleGetPersistedQueryList(w, request)
	default:
		s.t.Fatalf("unexpected operation name in data source test server: %s", request.OperationName)
	}
}

func (s *metadataDataSourceTestServer) handleGetGraphMetadata(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	if graphID != "inventory" {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": nil,
			},
		})
		return
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"id":    "inventory",
				"name":  "Inventory Graph",
				"title": "Inventory",
				"variants": []map[string]any{
					{"name": "current"},
					{"name": "staging"},
				},
			},
		},
	})
}

func (s *metadataDataSourceTestServer) handleGetVariantMetadata(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variantName := variableString(request.Variables, "variantName")
	if graphID != "inventory" {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": nil,
			},
		})
		return
	}

	var payload any
	switch variantName {
	case "current":
		payload = map[string]any{
			"id":   "variant-current",
			"name": "current",
			"subgraphs": []map[string]any{
				{
					"name":     "products",
					"url":      "https://products.example.com/graphql",
					"revision": "rev-products-1",
				},
				{
					"name":     "reviews",
					"url":      "https://reviews.example.com/graphql",
					"revision": "rev-reviews-2",
				},
			},
		}
	case "staging":
		payload = map[string]any{
			"id":   "variant-staging",
			"name": "staging",
			"subgraphs": []map[string]any{
				{
					"name":     "products",
					"url":      "https://staging-products.example.com/graphql",
					"revision": "rev-products-staging",
				},
			},
		}
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"id":      "inventory",
				"variant": payload,
			},
		},
	})
}

func (s *metadataDataSourceTestServer) handleGetPersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	id := variableString(request.Variables, "id")
	if graphID != "inventory" {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": nil,
			},
		})
		return
	}

	var payload any
	if id == "pql-123" {
		payload = map[string]any{
			"id":   "pql-123",
			"name": "mobile-clients",
			"linkedVariants": []map[string]any{
				{
					"name": "current",
					"graph": map[string]any{
						"id": "inventory",
					},
				},
				{
					"name": "staging",
					"graph": map[string]any{
						"id": "inventory",
					},
				},
			},
		}
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"persistedQueryList": payload,
			},
		},
	})
}
