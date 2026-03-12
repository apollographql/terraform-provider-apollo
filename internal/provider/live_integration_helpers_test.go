// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type liveIntegrationEnv struct {
	APIKey   string
	Endpoint string
	GraphID  string
}

type liveVariantSubgraphFixture struct {
	Env        liveIntegrationEnv
	Client     *Client
	GraphID    string
	Variant    string
	Subgraph   string
	RoutingURL string
	Revision   string
}

func requireLiveIntegrationEnv(t *testing.T, requireMutations bool) liveIntegrationEnv {
	t.Helper()

	if os.Getenv("APOLLO_PLATFORM_INTEGRATION") != "1" {
		t.Skip("set APOLLO_PLATFORM_INTEGRATION=1 to run live GraphOS integration tests")
	}

	if requireMutations && os.Getenv("APOLLO_INTEGRATION_ALLOW_MUTATIONS") != "1" {
		t.Skip("set APOLLO_INTEGRATION_ALLOW_MUTATIONS=1 to run mutating live GraphOS integration tests")
	}

	apiKey := strings.TrimSpace(os.Getenv(apiKeyEnvVar))
	graphID := strings.TrimSpace(os.Getenv("APOLLO_TEST_GRAPH_ID"))
	if apiKey == "" || graphID == "" {
		t.Skip("set APOLLO_KEY and APOLLO_TEST_GRAPH_ID to run live GraphOS integration tests")
	}

	endpoint := strings.TrimSpace(os.Getenv(endpointEnvVar))
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	return liveIntegrationEnv{
		APIKey:   apiKey,
		Endpoint: endpoint,
		GraphID:  graphID,
	}
}

func newLiveIntegrationClient(env liveIntegrationEnv) *Client {
	return NewClient(env.APIKey, env.Endpoint, "terraform-provider-apollo/live-integration")
}

func liveIntegrationContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 90*time.Second)
}

func uniqueLiveIntegrationName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func bootstrapLiveVariantSubgraphFixture(t *testing.T, env liveIntegrationEnv) *liveVariantSubgraphFixture {
	t.Helper()

	fixture := &liveVariantSubgraphFixture{
		Env:        env,
		Client:     newLiveIntegrationClient(env),
		GraphID:    env.GraphID,
		Variant:    uniqueLiveIntegrationName("tfp-variant"),
		Subgraph:   uniqueLiveIntegrationName("tfp-subgraph"),
		Revision:   uniqueLiveIntegrationName("tfp-revision"),
		RoutingURL: fmt.Sprintf("https://%s.example.invalid/graphql", uniqueLiveIntegrationName("tfp-router")),
	}

	ctx, cancel := liveIntegrationContext(t)
	defer cancel()

	if err := fixture.Client.PublishSubgraphSchema(ctx, PublishSubgraphInput{
		GraphID:    fixture.GraphID,
		Variant:    fixture.Variant,
		Subgraph:   fixture.Subgraph,
		RoutingURL: fixture.RoutingURL,
		SchemaSDL:  "type Query { integrationSmoke: String! }",
		Revision:   fixture.Revision,
	}); err != nil {
		t.Fatalf("bootstrap publishSubgraph failed: %s", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := liveIntegrationContext(t)
		defer cleanupCancel()

		if err := fixture.Client.RemoveSubgraph(cleanupCtx, fixture.GraphID, fixture.Variant, fixture.Subgraph); err != nil && !IsErrorKind(err, ErrorKindNotFound) {
			t.Logf("live cleanup: remove subgraph %s/%s/%s: %s", fixture.GraphID, fixture.Variant, fixture.Subgraph, err)
		}

		if err := fixture.Client.DeleteGraphVariant(cleanupCtx, fixture.GraphID, fixture.Variant); err != nil && !IsErrorKind(err, ErrorKindNotFound) {
			t.Logf("live cleanup: delete variant %s/%s: %s", fixture.GraphID, fixture.Variant, err)
		}
	})

	return fixture
}

func testCheckCaptureResourceAttr(resourceName string, attribute string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		value, ok := resourceState.Primary.Attributes[attribute]
		if !ok {
			return fmt.Errorf("attribute %q not found on resource %q", attribute, resourceName)
		}

		*target = value
		return nil
	}
}

func liveProviderConfig() string {
	return `
provider "apollo" {}
`
}

func liveMutableResourcesConfig(graphID string, variant string, subgraphName string, pqlName string, description string, graphKeyName string, graphKeyRole string, subgraphKeyName string) string {
	return fmt.Sprintf(`
%s

resource "apollo_persisted_query_list" "test" {
  graph_id    = %q
  name        = %q
  description = %q
}

resource "apollo_persisted_query_list_link" "test" {
  persisted_query_list_id = apollo_persisted_query_list.test.id
  graph_id                = apollo_persisted_query_list.test.graph_id
  variant                 = %q
}

resource "apollo_graph_api_key" "test" {
  graph_id = %q
  name     = %q
  role     = %q
}

resource "apollo_subgraph_api_key" "test" {
  graph_id      = %q
  variant       = %q
  subgraph_name = %q
  name          = %q
}
`, liveProviderConfig(), graphID, pqlName, description, variant, graphID, graphKeyName, graphKeyRole, graphID, variant, subgraphName, subgraphKeyName)
}

func liveSubgraphResourceConfig(graphID string, variant string, subgraphName string) string {
	return fmt.Sprintf(`
%s

resource "apollo_subgraph" "test" {
  graph_id = %q
  variant  = %q
  name     = %q
}
`, liveProviderConfig(), graphID, variant, subgraphName)
}

func liveGraphVariantResourceConfig(graphID string, variant string) string {
	return fmt.Sprintf(`
%s

resource "apollo_graph_variant" "test" {
  graph_id = %q
  variant  = %q
}
`, liveProviderConfig(), graphID, variant)
}
