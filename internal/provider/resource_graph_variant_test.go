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

func TestAccGraphVariantResourceDeferredCreate(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccGraphVariantResourceConfig(server.URL(), "production"),
				ExpectError: regexp.MustCompile(`Create graph variant: capability deferred`),
			},
		},
	})
}

func TestAccGraphVariantResourceImportReadDelete(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      "inventory:current",
				ImportStatePersist: true,
			},
			{
				Config: testAccGraphVariantResourceConfig(server.URL(), "current"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "graph_id", "inventory"),
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "variant", "current"),
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "id", "inventory@current"),
				),
			},
		},
	})

	if got := server.VariantCount(); got != 0 {
		t.Fatalf("expected imported variant to be deleted during destroy, got %d remaining", got)
	}
}

func TestAccGraphVariantResourceRefreshMissingVariant(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      "inventory:current",
				ImportStatePersist: true,
			},
			{
				PreConfig:          func() { server.DeleteVariant("inventory", "current") },
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGraphVariantResourceReadUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")
	server.SetMode(operationGetVariantMetadata, testServerModeUnauthenticated)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:  "apollo_graph_variant.test",
				ImportState:   true,
				ImportStateId: "inventory:current",
				ExpectError:   regexp.MustCompile(`Read graph variant: authentication failed`),
			},
		},
	})
}

func TestAccGraphVariantResourceReadPermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")
	server.SetMode(operationGetVariantMetadata, testServerModePermissionDenied)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:        testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:  "apollo_graph_variant.test",
				ImportState:   true,
				ImportStateId: "inventory:current",
				ExpectError:   regexp.MustCompile(`(?s)Read graph variant: permission denied.*Graph read or broader org access.*variant metadata`),
			},
		},
	})
}

func TestAccGraphVariantResourceDeleteMissingVariant(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      "inventory:current",
				ImportStatePersist: true,
			},
			{
				Config:    testAccGraphVariantResourceConfig(server.URL(), "current"),
				PreConfig: func() { server.DeleteVariant("inventory", "current") },
				Destroy:   true,
			},
		},
	})
}

func TestAccGraphVariantResourceDeleteUnauthenticated(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      "inventory:current",
				ImportStatePersist: true,
			},
			{
				Config:      testAccGraphVariantResourceConfig(server.URL(), "current"),
				PreConfig:   func() { server.SetMode(operationDeleteGraphVariant, testServerModeUnauthenticated) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Delete graph variant: authentication failed`),
			},
		},
	})
}

func TestAccGraphVariantResourceDeletePermissionDenied(t *testing.T) {
	requireHermeticAcceptance(t)

	server := newGraphVariantTestServer(t)
	defer server.Close()
	server.UpsertVariant("inventory", "current", "inventory@current")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGraphVariantResourceConfig(server.URL(), "current"),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      "inventory:current",
				ImportStatePersist: true,
			},
			{
				Config:      testAccGraphVariantResourceConfig(server.URL(), "current"),
				PreConfig:   func() { server.SetMode(operationDeleteGraphVariant, testServerModePermissionDenied) },
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)Delete graph variant: permission denied.*Graph admin or broader org access.*variant deletion`),
			},
		},
	})
}

func TestGraphVariantResourceModelFromAPI(t *testing.T) {
	t.Parallel()

	model := graphVariantResourceModelFromAPI(&VariantMetadata{
		ID:      " inventory@current ",
		GraphID: " inventory ",
		Name:    " current ",
	})

	if got, want := model.ID.ValueString(), "inventory@current"; got != want {
		t.Fatalf("unexpected id: got %q want %q", got, want)
	}

	if got, want := model.GraphID.ValueString(), "inventory"; got != want {
		t.Fatalf("unexpected graph id: got %q want %q", got, want)
	}

	if got, want := model.Variant.ValueString(), "current"; got != want {
		t.Fatalf("unexpected variant: got %q want %q", got, want)
	}
}

func TestGraphVariantResourceUpdateProviderBug(t *testing.T) {
	t.Parallel()

	resourceUnderTest := &graphVariantResource{}
	var resp tfresource.UpdateResponse

	resourceUnderTest.Update(context.Background(), tfresource.UpdateRequest{}, &resp)

	if !diagnosticsContainSummary(resp.Diagnostics, "Update graph variant: provider bug") {
		t.Fatalf("expected provider bug diagnostic, got %#v", resp.Diagnostics)
	}
}

func testAccGraphVariantResourceConfig(endpoint string, variant string) string {
	return fmt.Sprintf(`
provider "apollo" {
  api_key  = "service:key"
  endpoint = %q
}

resource "apollo_graph_variant" "test" {
  graph_id = "inventory"
  variant  = %q
}
`, endpoint, variant)
}

type graphVariantRecord struct {
	GraphID string
	ID      string
	Name    string
}

type graphVariantTestServer struct {
	t      *testing.T
	server *httptest.Server

	mu       sync.Mutex
	variants map[string]graphVariantRecord
	modes    map[string]testServerMode
}

func newGraphVariantTestServer(t *testing.T) *graphVariantTestServer {
	t.Helper()

	testServer := &graphVariantTestServer{
		t:        t,
		variants: map[string]graphVariantRecord{},
		modes:    map[string]testServerMode{},
	}

	testServer.server = httptest.NewServer(http.HandlerFunc(testServer.handle))
	return testServer
}

func (s *graphVariantTestServer) Close() {
	s.server.Close()
}

func (s *graphVariantTestServer) URL() string {
	return s.server.URL
}

func (s *graphVariantTestServer) SetMode(operation string, mode testServerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modes[operation] = mode
}

func (s *graphVariantTestServer) UpsertVariant(graphID string, variant string, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variants[graphVariantKey(graphID, variant)] = graphVariantRecord{
		GraphID: graphID,
		ID:      id,
		Name:    variant,
	}
}

func (s *graphVariantTestServer) DeleteVariant(graphID string, variant string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.variants, graphVariantKey(graphID, variant))
}

func (s *graphVariantTestServer) VariantCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.variants)
}

func (s *graphVariantTestServer) modeForOperation(operation string) testServerMode {
	s.mu.Lock()
	defer s.mu.Unlock()

	mode := s.modes[operation]
	if mode != testServerModeNormal {
		delete(s.modes, operation)
	}

	return mode
}

func (s *graphVariantTestServer) handle(w http.ResponseWriter, r *http.Request) {
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
	} else if mode == testServerModePermissionDenied {
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
	case operationDeleteGraphVariant:
		s.handleDeleteGraphVariant(w, request)
	default:
		s.t.Fatalf("unexpected operation name in graph variant test server: %s", request.OperationName)
	}
}

func (s *graphVariantTestServer) handleGetVariantMetadata(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variantName := variableString(request.Variables, "variantName")

	s.mu.Lock()
	record, ok := s.variants[graphVariantKey(graphID, variantName)]
	s.mu.Unlock()

	if !ok {
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
					"id":        record.ID,
					"name":      record.Name,
					"subgraphs": []map[string]any{},
				},
			},
		},
	})
}

func (s *graphVariantTestServer) handleDeleteGraphVariant(w http.ResponseWriter, request graphQLRequest) {
	graphID := variableString(request.Variables, "graphId")
	variantName := variableString(request.Variables, "variantName")

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.variants[graphVariantKey(graphID, variantName)]; !ok {
		writeJSONResponse(s.t, w, map[string]any{
			"data": map[string]any{
				"graph": map[string]any{
					"variant": nil,
				},
			},
		})
		return
	}

	delete(s.variants, graphVariantKey(graphID, variantName))
	writeJSONResponse(s.t, w, map[string]any{
		"data": map[string]any{
			"graph": map[string]any{
				"variant": map[string]any{
					"delete": map[string]any{
						"deleted": true,
					},
				},
			},
		},
	})
}

func graphVariantKey(graphID string, variant string) string {
	return JoinImportID(graphID, variant)
}
