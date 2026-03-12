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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGraphAPIKeyResourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "name", "inventory-ci"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "role", "GRAPH_ADMIN"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "key", "service:graph-token-1"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "id", "graph-key-1"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "created_at", "2026-03-11T00:00:00Z"),
					testCheckNullOrMissingResourceAttr("apollo_graph_api_key.test", "last_used_at"),
				),
			},
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-deploy", "GRAPH_ADMIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "name", "inventory-deploy"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "role", "GRAPH_ADMIN"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "key", "service:graph-token-1"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "id", "graph-key-1"),
				),
			},
		},
	})

	if got := server.GraphKeyCount(); got != 0 {
		t.Fatalf("expected graph API keys to be destroyed after test, got %d", got)
	}
}

func TestAccGraphAPIKeyResourceImport(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()
	server.UpsertGraphAPIKey(graphAPIKeyRecord{
		GraphID:    "inventory",
		ID:         "graph-key-existing",
		Name:       "inventory-ci",
		Role:       "GRAPH_ADMIN",
		Token:      "service:graph-token-existing",
		CreatedAt:  "2026-03-11T00:00:00Z",
		LastUsedAt: "",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				ResourceName:       "apollo_graph_api_key.test",
				ImportState:        true,
				ImportStateId:      "inventory:graph-key-existing",
				ImportStatePersist: true,
			},
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "name", "inventory-ci"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "role", "GRAPH_ADMIN"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "id", "graph-key-existing"),
					testCheckNullOrMissingResourceAttr("apollo_graph_api_key.test", "key"),
				),
			},
		},
	})

	if got := server.GraphKeyCount(); got != 0 {
		t.Fatalf("expected imported graph API key to be destroyed after test, got %d", got)
	}
}

func TestAccGraphAPIKeyResourceRoleReplacement(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "id", "graph-key-1"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "key", "service:graph-token-1"),
				),
			},
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "CONTRIBUTOR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "role", "CONTRIBUTOR"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "id", "graph-key-2"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "key", "service:graph-token-2"),
				),
			},
		},
	})
}

func TestAccGraphAPIKeyResourceRefreshMissingRemote(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
			},
			{
				PreConfig:          func() { server.DeleteGraphAPIKeyRecord("inventory", "graph-key-1") },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGraphAPIKeyResourceReadPermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()
	server.UpsertGraphAPIKey(graphAPIKeyRecord{
		GraphID:    "inventory",
		ID:         "graph-key-existing",
		Name:       "inventory-ci",
		Role:       "GRAPH_ADMIN",
		Token:      "service:graph-token-existing",
		CreatedAt:  "2026-03-11T00:00:00Z",
		LastUsedAt: "",
	})
	server.SetMode(operationGetGraphAPIKey, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				ResourceName:  "apollo_graph_api_key.test",
				ImportState:   true,
				ImportStateId: "inventory:graph-key-existing",
				ExpectError:   regexp.MustCompile(`(?s)Read graph API key: permission denied.*Org Admin or Graph Admin access.*graph API key management`),
			},
		},
	})
}

func TestAccGraphAPIKeyResourceDeleteUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
			},
			{
				Config:      testAccGraphAPIKeyResourceConfig(server.URL(), "inventory", "inventory-ci", "GRAPH_ADMIN"),
				PreConfig:   func() { server.SetMode(operationDeleteGraphAPIKey, testServerModeUnauthenticated) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Delete graph API key: authentication failed`),
			},
		},
	})
}

func TestAccSubgraphAPIKeyResourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "variant", "production"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "subgraph_name", "products"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "name", "products-router"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "id", "subgraph-key-1"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "key", "service:subgraph-token-1"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "created_at", "2026-03-11T00:00:00Z"),
					testCheckNullOrMissingResourceAttr("apollo_subgraph_api_key.test", "last_used_at"),
				),
			},
			{
				Config: testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-runtime"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "name", "products-runtime"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "id", "subgraph-key-1"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "key", "service:subgraph-token-1"),
				),
			},
		},
	})

	if got := server.SubgraphKeyCount(); got != 0 {
		t.Fatalf("expected subgraph API keys to be destroyed after test, got %d", got)
	}
}

func TestAccSubgraphAPIKeyResourceImport(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()
	server.UpsertSubgraphAPIKey(subgraphAPIKeyRecord{
		OrganizationID: "org-1",
		ID:             "subgraph-key-existing",
		Name:           "products-router",
		Token:          "service:subgraph-token-existing",
		CreatedAt:      "2026-03-11T00:00:00Z",
		Targets: []subgraphAPIKeyTarget{
			{GraphID: "inventory", Variant: "production", SubgraphName: "products"},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				ResourceName:       "apollo_subgraph_api_key.test",
				ImportState:        true,
				ImportStateId:      "inventory:production:products:subgraph-key-existing",
				ImportStatePersist: true,
			},
			{
				Config: testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "variant", "production"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "subgraph_name", "products"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "name", "products-router"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "id", "subgraph-key-existing"),
					testCheckNullOrMissingResourceAttr("apollo_subgraph_api_key.test", "key"),
				),
			},
		},
	})

	if got := server.SubgraphKeyCount(); got != 0 {
		t.Fatalf("expected imported subgraph API key to be destroyed after test, got %d", got)
	}
}

func TestAccSubgraphAPIKeyResourceRefreshMissingRemote(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
			},
			{
				PreConfig:          func() { server.DeleteSubgraphAPIKeyRecord("subgraph-key-1") },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSubgraphAPIKeyResourceReadUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()
	server.UpsertSubgraphAPIKey(subgraphAPIKeyRecord{
		OrganizationID: "org-1",
		ID:             "subgraph-key-existing",
		Name:           "products-router",
		Token:          "service:subgraph-token-existing",
		CreatedAt:      "2026-03-11T00:00:00Z",
		Targets: []subgraphAPIKeyTarget{
			{GraphID: "inventory", Variant: "production", SubgraphName: "products"},
		},
	})
	server.SetMode(operationGetSubgraphAPIKey, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				ResourceName:  "apollo_subgraph_api_key.test",
				ImportState:   true,
				ImportStateId: "inventory:production:products:subgraph-key-existing",
				ExpectError:   regexp.MustCompile(`Read subgraph API key: authentication failed`),
			},
		},
	})
}

func TestAccSubgraphAPIKeyResourceDeletePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
			},
			{
				Config:      testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				PreConfig:   func() { server.SetMode(operationDeleteSubgraphAPIKey, testServerModePermissionDenied) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)Delete subgraph API key: permission denied.*Org Admin or Graph Admin access.*subgraph API key management`),
			},
		},
	})
}

func TestAccSubgraphAPIKeyResourceRejectsMultiTargetRemoteKey(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newAPIKeyTestServer(t)
	defer server.Close()
	server.UpsertSubgraphAPIKey(subgraphAPIKeyRecord{
		OrganizationID: "org-1",
		ID:             "subgraph-key-multi",
		Name:           "products-router",
		CreatedAt:      "2026-03-11T00:00:00Z",
		Targets: []subgraphAPIKeyTarget{
			{GraphID: "inventory", Variant: "production", SubgraphName: "products"},
			{GraphID: "inventory", Variant: "staging", SubgraphName: "products"},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccSubgraphAPIKeyResourceConfig(server.URL(), "inventory", "production", "products", "products-router"),
				ResourceName:  "apollo_subgraph_api_key.test",
				ImportState:   true,
				ImportStateId: "inventory:production:products:subgraph-key-multi",
				ExpectError:   regexp.MustCompile(`Read subgraph API key: capability deferred`),
			},
		},
	})
}

func testAccGraphAPIKeyResourceConfig(endpoint string, graphID string, name string, role string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

resource "apollo_graph_api_key" "test" {
  graph_id = %q
  name     = %q
  role     = %q
}
`, endpoint, graphID, name, role)
}

func testAccSubgraphAPIKeyResourceConfig(endpoint string, graphID string, variant string, subgraphName string, name string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

resource "apollo_subgraph_api_key" "test" {
  graph_id      = %q
  variant       = %q
  subgraph_name = %q
  name          = %q
}
`, endpoint, graphID, variant, subgraphName, name)
}

func testCheckNullOrMissingResourceAttr(resourceName string, attribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		if value, ok := resourceState.Primary.Attributes[attribute]; ok && value != "" {
			return fmt.Errorf("expected attribute %q to be null or absent, got %q", attribute, value)
		}

		return nil
	}
}

type graphAPIKeyRecord struct {
	GraphID    string
	ID         string
	Name       string
	Role       string
	Token      string
	CreatedAt  string
	LastUsedAt string
}

type subgraphAPIKeyTarget struct {
	GraphID      string
	Variant      string
	SubgraphName string
}

type subgraphAPIKeyRecord struct {
	OrganizationID string
	ID             string
	Name           string
	Token          string
	CreatedAt      string
	LastUsedAt     string
	Targets        []subgraphAPIKeyTarget
}

type apiKeyTestServer struct {
	t      *testing.T
	server *httptest.Server

	mu                 sync.Mutex
	graphKeys          map[string]graphAPIKeyRecord
	subgraphKeys       map[string]subgraphAPIKeyRecord
	graphOrganizations map[string]string
	modes              map[string]testServerMode
	graphKeyCounter    int
	subgraphKeyCounter int
}

func newAPIKeyTestServer(t *testing.T) *apiKeyTestServer {
	t.Helper()

	testServer := &apiKeyTestServer{
		t:                  t,
		graphKeys:          map[string]graphAPIKeyRecord{},
		subgraphKeys:       map[string]subgraphAPIKeyRecord{},
		graphOrganizations: map[string]string{"inventory": "org-1"},
		modes:              map[string]testServerMode{},
	}

	testServer.server = httptest.NewServer(http.HandlerFunc(testServer.handle))
	return testServer
}

func (s *apiKeyTestServer) Close() {
	s.server.Close()
}

func (s *apiKeyTestServer) URL() string {
	return s.server.URL
}

func (s *apiKeyTestServer) SetMode(operation string, mode testServerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modes[operation] = mode
}

func (s *apiKeyTestServer) UpsertGraphAPIKey(record graphAPIKeyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphKeys[JoinImportID(record.GraphID, record.ID)] = record
}

func (s *apiKeyTestServer) DeleteGraphAPIKeyRecord(graphID string, keyID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.graphKeys, JoinImportID(graphID, keyID))
}

func (s *apiKeyTestServer) GraphKeyCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.graphKeys)
}

func (s *apiKeyTestServer) UpsertSubgraphAPIKey(record subgraphAPIKeyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subgraphKeys[record.ID] = record
}

func (s *apiKeyTestServer) DeleteSubgraphAPIKeyRecord(keyID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subgraphKeys, keyID)
}

func (s *apiKeyTestServer) SubgraphKeyCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subgraphKeys)
}

func (s *apiKeyTestServer) modeForOperation(operation string) testServerMode {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := s.modes[operation]
	if mode != testServerModeNormal {
		delete(s.modes, operation)
	}

	return mode
}

func (s *apiKeyTestServer) handle(w http.ResponseWriter, r *http.Request) {
	s.t.Helper()

	var request graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.t.Fatalf("failed to decode GraphQL request: %s", err)
	}

	switch s.modeForOperation(request.OperationName) {
	case testServerModeUnauthenticated:
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(s.t, w, graphQLErrorPayload("unauthorized", "UNAUTHENTICATED", http.StatusUnauthorized))
		return
	case testServerModePermissionDenied:
		w.WriteHeader(http.StatusForbidden)
		writeJSONResponse(s.t, w, graphQLErrorPayload("permission denied", "PERMISSION_DENIED", http.StatusForbidden))
		return
	}

	switch request.OperationName {
	case operationCreateGraphAPIKey:
		s.handleCreateGraphAPIKey(w, request)
	case operationGetGraphAPIKey:
		s.handleGetGraphAPIKey(w, request)
	case operationUpdateGraphAPIKey:
		s.handleUpdateGraphAPIKey(w, request)
	case operationDeleteGraphAPIKey:
		s.handleDeleteGraphAPIKey(w, request)
	case operationResolveGraphOrgID:
		s.handleResolveGraphOrganizationID(w, request)
	case operationCreateSubgraphAPIKey:
		s.handleCreateSubgraphAPIKey(w, request)
	case operationGetSubgraphAPIKey:
		s.handleGetSubgraphAPIKey(w, request)
	case operationUpdateSubgraphAPIKey:
		s.handleUpdateSubgraphAPIKey(w, request)
	case operationDeleteSubgraphAPIKey:
		s.handleDeleteSubgraphAPIKey(w, request)
	default:
		s.t.Fatalf("unexpected operation name in API key test server: %s", request.OperationName)
	}
}

func (s *apiKeyTestServer) handleCreateGraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	name := variableString(request.Variables, "name")
	role := variableString(request.Variables, "role")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.graphKeyCounter++
	record := graphAPIKeyRecord{
		GraphID:    graphID,
		ID:         fmt.Sprintf("graph-key-%d", s.graphKeyCounter),
		Name:       name,
		Role:       role,
		Token:      fmt.Sprintf("service:graph-token-%d", s.graphKeyCounter),
		CreatedAt:  "2026-03-11T00:00:00Z",
		LastUsedAt: "",
	}
	s.graphKeys[JoinImportID(record.GraphID, record.ID)] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"newKey": graphAPIKeyRecordPayload(record, true),
			},
		},
	})
}

func (s *apiKeyTestServer) handleGetGraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]map[string]any, 0)
	for _, record := range s.graphKeys {
		if record.GraphID == graphID {
			keys = append(keys, graphAPIKeyRecordPayload(record, false))
		}
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"id":      graphID,
				"apiKeys": keys,
			},
		},
	})
}

func (s *apiKeyTestServer) handleUpdateGraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	keyID := variableString(request.Variables, "id")
	name := variableString(request.Variables, "name")

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.graphKeys[JoinImportID(graphID, keyID)]
	if !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"renameKey": nil,
				},
			},
		})
		return
	}

	record.Name = name
	s.graphKeys[JoinImportID(graphID, keyID)] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"renameKey": graphAPIKeyRecordPayload(record, false),
			},
		},
	})
}

func (s *apiKeyTestServer) handleDeleteGraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	keyID := variableString(request.Variables, "id")

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.graphKeys, JoinImportID(graphID, keyID))

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"removeKey": nil,
			},
		},
	})
}

func (s *apiKeyTestServer) handleResolveGraphOrganizationID(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")

	s.mu.Lock()
	defer s.mu.Unlock()

	orgID := s.graphOrganizations[graphID]
	if orgID == "" {
		orgID = "org-1"
		s.graphOrganizations[graphID] = orgID
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"id": graphID,
				"account": map[string]any{
					"id": orgID,
				},
			},
		},
	})
}

func (s *apiKeyTestServer) handleCreateSubgraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	organizationID := variableString(request.Variables, "organizationId")
	name := variableString(request.Variables, "name")

	resources, _ := request.Variables["resources"].(map[string]any)
	subgraphs, _ := resources["subgraphs"].([]any)
	if len(subgraphs) != 1 {
		s.t.Fatalf("expected one subgraph target, got %#v", subgraphs)
	}

	targetMap, _ := subgraphs[0].(map[string]any)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subgraphKeyCounter++
	record := subgraphAPIKeyRecord{
		OrganizationID: organizationID,
		ID:             fmt.Sprintf("subgraph-key-%d", s.subgraphKeyCounter),
		Name:           name,
		Token:          fmt.Sprintf("service:subgraph-token-%d", s.subgraphKeyCounter),
		CreatedAt:      "2026-03-11T00:00:00Z",
		LastUsedAt:     "",
		Targets: []subgraphAPIKeyTarget{
			{
				GraphID:      variableString(targetMap, "graphId"),
				Variant:      variableString(targetMap, "variantName"),
				SubgraphName: variableString(targetMap, "subgraphName"),
			},
		},
	}
	s.subgraphKeys[record.ID] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"organization": map[string]any{
				"createKey": subgraphAPIKeyRecordPayload(record, true),
			},
		},
	})
}

func (s *apiKeyTestServer) handleGetSubgraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	organizationID := variableString(request.Variables, "organizationId")
	keyID := variableString(request.Variables, "keyId")

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.subgraphKeys[keyID]
	if !ok || record.OrganizationID != organizationID {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"organization": map[string]any{
					"apiKey": nil,
				},
			},
		})
		return
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"organization": map[string]any{
				"apiKey": subgraphAPIKeyRecordPayload(record, false),
			},
		},
	})
}

func (s *apiKeyTestServer) handleUpdateSubgraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	organizationID := variableString(request.Variables, "organizationId")
	keyID := variableString(request.Variables, "keyId")
	name := variableString(request.Variables, "name")

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.subgraphKeys[keyID]
	if !ok || record.OrganizationID != organizationID {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"organization": map[string]any{
					"renameKey": nil,
				},
			},
		})
		return
	}

	record.Name = name
	s.subgraphKeys[keyID] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"organization": map[string]any{
				"renameKey": subgraphAPIKeyRecordPayload(record, false),
			},
		},
	})
}

func (s *apiKeyTestServer) handleDeleteSubgraphAPIKey(w http.ResponseWriter, request graphQLRequest) {
	keyID := variableString(request.Variables, "keyId")

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subgraphKeys, keyID)

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"organization": map[string]any{
				"deleteKey": keyID,
			},
		},
	})
}

func graphAPIKeyRecordPayload(record graphAPIKeyRecord, includeToken bool) map[string]any {
	payload := map[string]any{
		"id":        record.ID,
		"keyName":   record.Name,
		"role":      record.Role,
		"createdAt": record.CreatedAt,
		"lastUsed":  nil,
	}

	if normalized := NormalizeString(record.LastUsedAt); normalized != "" {
		payload["lastUsed"] = normalized
	}

	if includeToken {
		payload["token"] = record.Token
	}

	return payload
}

func subgraphAPIKeyRecordPayload(record subgraphAPIKeyRecord, includeToken bool) map[string]any {
	resources := make([]map[string]any, 0, len(record.Targets))
	for _, target := range record.Targets {
		resources = append(resources, map[string]any{
			"resourceId":   JoinImportID(target.GraphID, target.Variant, target.SubgraphName),
			"resourceType": "SUBGRAPH",
		})
	}

	payload := map[string]any{
		"id":        record.ID,
		"keyName":   record.Name,
		"keyType":   "SUBGRAPH",
		"createdAt": record.CreatedAt,
		"lastUsed":  nil,
		"resources": resources,
	}

	if normalized := NormalizeString(record.LastUsedAt); normalized != "" {
		payload["lastUsed"] = normalized
	}

	if includeToken {
		payload["token"] = record.Token
	}

	return payload
}
