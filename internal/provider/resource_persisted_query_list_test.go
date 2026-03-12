// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPersistedQueryListResourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newPersistedQueryListTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPersistedQueryListResourceConfig(server.URL(), "Mobile clients"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "name", "mobile-clients"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "description", "Mobile clients"),
					resource.TestCheckResourceAttrSet("apollo_persisted_query_list.test", "id"),
				),
			},
			{
				Config: testAccPersistedQueryListResourceConfig(server.URL(), "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "name", "mobile-clients"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "description", "Updated description"),
				),
			},
			{
				ResourceName: "apollo_persisted_query_list.test",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState, ok := state.RootModule().Resources["apollo_persisted_query_list.test"]
					if !ok {
						return "", fmt.Errorf("persisted query list state missing")
					}

					return JoinImportID(resourceState.Primary.Attributes["graph_id"], resourceState.Primary.Attributes["id"]), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})

	if got := server.CollectionCount(); got != 0 {
		t.Fatalf("expected collections to be destroyed after test, got %d", got)
	}
}

func TestAccPersistedQueryListResourceUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newPersistedQueryListTestServer(t)
	defer server.Close()
	server.SetMode(operationCreatePersistedQueryList, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccPersistedQueryListResourceConfig(server.URL(), "Mobile clients"),
				ExpectError: regexp.MustCompile(`Create persisted query list: authentication failed`),
			},
		},
	})
}

func TestAccPersistedQueryListResourcePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newPersistedQueryListTestServer(t)
	defer server.Close()
	server.SetMode(operationCreatePersistedQueryList, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccPersistedQueryListResourceConfig(server.URL(), "Mobile clients"),
				ExpectError: regexp.MustCompile(`Create persisted query list: permission denied`),
			},
		},
	})
}

func TestAccPersistedQueryListLinkResourceBasic(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newPersistedQueryListTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPersistedQueryListLinkResourceConfig(server.URL(), "current"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "variant", "current"),
					resource.TestCheckResourceAttrPair("apollo_persisted_query_list_link.test", "persisted_query_list_id", "apollo_persisted_query_list.test", "id"),
				),
			},
			{
				Config: testAccPersistedQueryListLinkResourceConfig(server.URL(), "staging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "variant", "staging"),
					resource.TestCheckResourceAttrPair("apollo_persisted_query_list_link.test", "persisted_query_list_id", "apollo_persisted_query_list.test", "id"),
				),
			},
			{
				ResourceName:      "apollo_persisted_query_list_link.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	if got := server.LinkCount(); got != 0 {
		t.Fatalf("expected persisted query list links to be destroyed after test, got %d", got)
	}
}

func TestPersistedQueryListModelFromAPI(t *testing.T) {
	t.Parallel()

	model := persistedQueryListModelFromAPI(" inventory ", &PersistedQueryList{
		ID:   " pql-123 ",
		Name: " mobile-clients ",
	}, "   ")

	if got, want := model.ID.ValueString(), "pql-123"; got != want {
		t.Fatalf("unexpected id: got %q want %q", got, want)
	}

	if got, want := model.GraphID.ValueString(), "inventory"; got != want {
		t.Fatalf("unexpected graph id: got %q want %q", got, want)
	}

	if got, want := model.Name.ValueString(), "mobile-clients"; got != want {
		t.Fatalf("unexpected name: got %q want %q", got, want)
	}

	if !model.Description.IsNull() {
		t.Fatalf("expected empty description to normalize to null, got %v", model.Description)
	}
}

func TestPersistedQueryListHasVariantLink(t *testing.T) {
	t.Parallel()

	pql := &PersistedQueryList{
		ID:   "pql-123",
		Name: "mobile-clients",
		LinkedVariants: []GraphVariantRef{
			{GraphID: "inventory", Variant: "current"},
		},
	}

	if !persistedQueryListHasVariantLink(pql, "inventory", "current") {
		t.Fatal("expected current link to be present")
	}

	if persistedQueryListHasVariantLink(pql, "inventory", "staging") {
		t.Fatal("did not expect staging link to be present")
	}
}

func testAccPersistedQueryListResourceConfig(endpoint string, description string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key   = "service:key"
  endpoint  = %q
}

resource "apollo_persisted_query_list" "test" {
  graph_id    = "inventory"
  name        = "mobile-clients"
  description = %q
}
`, endpoint, description)
}

func testAccPersistedQueryListLinkResourceConfig(endpoint string, variant string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key   = "service:key"
  endpoint  = %q
}

resource "apollo_persisted_query_list" "test" {
  graph_id = "inventory"
  name     = "mobile-clients"
}

resource "apollo_persisted_query_list_link" "test" {
  persisted_query_list_id = apollo_persisted_query_list.test.id
  graph_id                = apollo_persisted_query_list.test.graph_id
  variant                 = %q
}
`, endpoint, variant)
}

type persistedQueryListRecord struct {
	GraphID     string
	ID          string
	Name        string
	Description string
}

type persistedQueryListTestServer struct {
	t      *testing.T
	server *httptest.Server

	mu          sync.Mutex
	nextID      int
	collections map[string]persistedQueryListRecord
	links       map[string]string
	modes       map[string]testServerMode
}

func newPersistedQueryListTestServer(t *testing.T) *persistedQueryListTestServer {
	t.Helper()

	testServer := &persistedQueryListTestServer{
		t:           t,
		nextID:      123,
		collections: map[string]persistedQueryListRecord{},
		links:       map[string]string{},
		modes:       map[string]testServerMode{},
	}

	testServer.server = httptest.NewServer(http.HandlerFunc(testServer.handle))
	return testServer
}

func (s *persistedQueryListTestServer) Close() {
	s.server.Close()
}

func (s *persistedQueryListTestServer) URL() string {
	return s.server.URL
}

func (s *persistedQueryListTestServer) SetMode(operation string, mode testServerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modes[operation] = mode
}

func (s *persistedQueryListTestServer) CollectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.collections)
}

func (s *persistedQueryListTestServer) LinkCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.links)
}

func (s *persistedQueryListTestServer) handle(w http.ResponseWriter, r *http.Request) {
	s.t.Helper()

	var request graphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.t.Fatalf("failed to decode GraphQL request: %s", err)
	}

	if mode := s.modeForOperation(request.OperationName); mode == testServerModeUnauthenticated {
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
	}

	switch request.OperationName {
	case operationCreatePersistedQueryList:
		s.handleCreatePersistedQueryList(w, request)
	case operationGetPersistedQueryList:
		s.handleGetPersistedQueryList(w, request)
	case operationUpdatePersistedQueryList:
		s.handleUpdatePersistedQueryList(w, request)
	case operationDeletePersistedQueryList:
		s.handleDeletePersistedQueryList(w, request)
	case operationGetVariantPersistedQueryList:
		s.handleGetVariantPersistedQueryList(w, request)
	case operationLinkPersistedQueryListVariant:
		s.handleLinkPersistedQueryListVariant(w, request)
	case operationUnlinkPersistedQueryList:
		s.handleUnlinkPersistedQueryListVariant(w, request)
	default:
		s.t.Fatalf("unexpected operation name in test server: %s", request.OperationName)
	}
}

func (s *persistedQueryListTestServer) handleCreatePersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	if s.modeForOperation(operationCreatePersistedQueryList) == testServerModePermissionDenied {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"createPersistedQueryList": map[string]any{
						"__typename": "PermissionError",
						"message":    "permission denied",
					},
				},
			},
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	graphID := variableString(request.Variables, "graphId")
	id := fmt.Sprintf("pql-%d", s.nextID)
	s.nextID++

	record := persistedQueryListRecord{
		GraphID:     graphID,
		ID:          id,
		Name:        variableString(request.Variables, "name"),
		Description: variableString(request.Variables, "description"),
	}
	s.collections[collectionKey(graphID, id)] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"createPersistedQueryList": map[string]any{
					"__typename":         "CreatePersistedQueryListResult",
					"persistedQueryList": s.collectionPayload(record),
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleGetPersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	id := variableString(request.Variables, "id")

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.collections[collectionKey(graphID, id)]
	if !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": nil,
				},
			},
		})
		return
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"persistedQueryList": s.collectionPayload(record),
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleUpdatePersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	if s.modeForOperation(operationUpdatePersistedQueryList) == testServerModePermissionDenied {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": map[string]any{
						"updateMetadata": map[string]any{
							"__typename": "PermissionError",
							"message":    "permission denied",
						},
					},
				},
			},
		})
		return
	}

	graphID := variableString(request.Variables, "graphId")
	id := variableString(request.Variables, "id")

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.collections[collectionKey(graphID, id)]
	if !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": nil,
				},
			},
		})
		return
	}

	record.Name = variableString(request.Variables, "name")
	record.Description = variableString(request.Variables, "description")
	s.collections[collectionKey(graphID, id)] = record

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"persistedQueryList": map[string]any{
					"updateMetadata": map[string]any{
						"__typename":         "UpdatePersistedQueryListMetadataResult",
						"persistedQueryList": s.collectionPayload(record),
					},
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleDeletePersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	if s.modeForOperation(operationDeletePersistedQueryList) == testServerModePermissionDenied {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": map[string]any{
						"delete": map[string]any{
							"__typename": "PermissionError",
							"message":    "permission denied",
						},
					},
				},
			},
		})
		return
	}

	graphID := variableString(request.Variables, "graphId")
	id := variableString(request.Variables, "id")

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.collections[collectionKey(graphID, id)]; !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": nil,
				},
			},
		})
		return
	}

	if s.isLinked(graphID, id) {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"persistedQueryList": map[string]any{
						"delete": map[string]any{
							"__typename": "CannotDeleteLinkedPersistedQueryListError",
							"message":    "persisted query list is still linked",
						},
					},
				},
			},
		})
		return
	}

	delete(s.collections, collectionKey(graphID, id))
	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"persistedQueryList": map[string]any{
					"delete": map[string]any{
						"__typename": "DeletePersistedQueryListResult",
						"graph": map[string]any{
							"id": graphID,
						},
					},
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleGetVariantPersistedQueryList(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variant := variableString(request.Variables, "variantName")

	s.mu.Lock()
	defer s.mu.Unlock()

	if !variantExists(graphID, variant) {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": nil,
				},
			},
		})
		return
	}

	pqlID := s.links[variantKey(graphID, variant)]
	var payload any
	if pqlID != "" {
		if record, ok := s.collections[collectionKey(graphID, pqlID)]; ok {
			payload = s.collectionPayload(record)
		}
	}

	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"variant": map[string]any{
					"persistedQueryList": payload,
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleLinkPersistedQueryListVariant(w http.ResponseWriter, request graphQLRequest) {
	if s.modeForOperation(operationLinkPersistedQueryListVariant) == testServerModePermissionDenied {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": map[string]any{
						"linkPersistedQueryList": map[string]any{
							"__typename": "PermissionError",
							"message":    "permission denied",
						},
					},
				},
			},
		})
		return
	}

	graphID := variableString(request.Variables, "graphId")
	variant := variableString(request.Variables, "variantName")
	pqlID := variableString(request.Variables, "persistedQueryListId")

	s.mu.Lock()
	defer s.mu.Unlock()

	if !variantExists(graphID, variant) {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": nil,
				},
			},
		})
		return
	}

	if _, ok := s.collections[collectionKey(graphID, pqlID)]; !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": map[string]any{
						"linkPersistedQueryList": map[string]any{
							"__typename": "ListNotFoundError",
							"message":    "persisted query list not found",
							"listId":     pqlID,
						},
					},
				},
			},
		})
		return
	}

	if existing := s.links[variantKey(graphID, variant)]; existing != "" {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": map[string]any{
						"linkPersistedQueryList": map[string]any{
							"__typename": "VariantAlreadyLinkedError",
							"message":    "variant already linked",
						},
					},
				},
			},
		})
		return
	}

	s.links[variantKey(graphID, variant)] = pqlID
	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"variant": map[string]any{
					"linkPersistedQueryList": map[string]any{
						"__typename": "LinkPersistedQueryListSuccess",
					},
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) handleUnlinkPersistedQueryListVariant(w http.ResponseWriter, request graphQLRequest) {
	if s.modeForOperation(operationUnlinkPersistedQueryList) == testServerModePermissionDenied {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": map[string]any{
						"unlinkPersistedQueryList": map[string]any{
							"__typename": "PermissionError",
							"message":    "permission denied",
						},
					},
				},
			},
		})
		return
	}

	graphID := variableString(request.Variables, "graphId")
	variant := variableString(request.Variables, "variantName")

	s.mu.Lock()
	defer s.mu.Unlock()

	if !variantExists(graphID, variant) {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": nil,
				},
			},
		})
		return
	}

	if s.links[variantKey(graphID, variant)] == "" {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": map[string]any{
						"unlinkPersistedQueryList": map[string]any{
							"__typename": "VariantAlreadyUnlinkedError",
							"message":    "variant already unlinked",
						},
					},
				},
			},
		})
		return
	}

	delete(s.links, variantKey(graphID, variant))
	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"variant": map[string]any{
					"unlinkPersistedQueryList": map[string]any{
						"__typename": "UnlinkPersistedQueryListSuccess",
					},
				},
			},
		},
	})
}

func (s *persistedQueryListTestServer) collectionPayload(record persistedQueryListRecord) map[string]any {
	linkedVariants := make([]map[string]any, 0, 1)
	for variantKey, linkedID := range s.links {
		if linkedID != record.ID {
			continue
		}

		graphID, variant := splitVariantKey(variantKey)
		if graphID != record.GraphID {
			continue
		}

		linkedVariants = append(linkedVariants, map[string]any{
			"name": variant,
			"graph": map[string]any{
				"id": graphID,
			},
		})
	}

	return map[string]any{
		"id":             record.ID,
		"name":           record.Name,
		"linkedVariants": linkedVariants,
	}
}

func (s *persistedQueryListTestServer) isLinked(graphID string, persistedQueryListID string) bool {
	for key, linkedID := range s.links {
		if linkedID != persistedQueryListID {
			continue
		}

		if variantGraphID, _ := splitVariantKey(key); variantGraphID == graphID {
			return true
		}
	}

	return false
}

func (s *persistedQueryListTestServer) modeForOperation(operation string) testServerMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modes[operation]
}

func collectionKey(graphID string, id string) string {
	return JoinImportID(graphID, id)
}

func variantKey(graphID string, variant string) string {
	return JoinImportID(graphID, variant)
}

func splitVariantKey(key string) (string, string) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func variantExists(graphID string, variant string) bool {
	return NormalizeString(graphID) != "" && NormalizeString(variant) != ""
}
