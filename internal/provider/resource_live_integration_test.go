// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestTerraformIntegrationMutableResources(t *testing.T) {
	env := requireLiveIntegrationEnv(t, true)
	fixture := bootstrapLiveVariantSubgraphFixture(t, env)

	var pqlID string
	var graphKeyID string
	var subgraphKeyID string

	initialPQLName := uniqueLiveIntegrationName("tfp-pql")
	initialGraphKeyName := uniqueLiveIntegrationName("tfp-graph-key")
	initialSubgraphKeyName := uniqueLiveIntegrationName("tfp-subgraph-key")

	updatedGraphKeyName := initialGraphKeyName + "-updated"
	updatedSubgraphKeyName := initialSubgraphKeyName + "-updated"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(_ *terraform.State) error {
			checkCtx, checkCancel := liveIntegrationContext(t)
			defer checkCancel()

			if pqlID != "" {
				_, err := fixture.Client.GetPersistedQueryList(checkCtx, fixture.GraphID, pqlID)
				if err == nil {
					return fmt.Errorf("persisted query list %q still exists after destroy", pqlID)
				}
				if !IsErrorKind(err, ErrorKindNotFound) {
					return fmt.Errorf("check persisted query list destroy: %w", err)
				}
			}

			if graphKeyID != "" {
				_, err := fixture.Client.GetGraphAPIKey(checkCtx, fixture.GraphID, graphKeyID)
				if err == nil {
					return fmt.Errorf("graph API key %q still exists after destroy", graphKeyID)
				}
				if !IsErrorKind(err, ErrorKindNotFound) {
					return fmt.Errorf("check graph API key destroy: %w", err)
				}
			}

			if subgraphKeyID != "" {
				_, err := fixture.Client.GetSubgraphAPIKey(checkCtx, SubgraphAPIKeyInput{
					GraphID:      fixture.GraphID,
					Variant:      fixture.Variant,
					SubgraphName: fixture.Subgraph,
					ID:           subgraphKeyID,
				})
				if err == nil {
					return fmt.Errorf("subgraph API key %q still exists after destroy", subgraphKeyID)
				}
				if !IsErrorKind(err, ErrorKindNotFound) && !IsErrorKind(err, ErrorKindPermissionDenied) && !strings.Contains(err.Error(), "NOT_ALLOWED_BY_USER_ROLE") {
					return fmt.Errorf("check subgraph API key destroy: %w", err)
				}
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: liveMutableResourcesConfig(
					fixture.GraphID,
					fixture.Variant,
					fixture.Subgraph,
					initialPQLName,
					"Live integration smoke collection",
					initialGraphKeyName,
					"GRAPH_ADMIN",
					initialSubgraphKeyName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckCaptureResourceAttr("apollo_persisted_query_list.test", "id", &pqlID),
					testCheckCaptureResourceAttr("apollo_graph_api_key.test", "id", &graphKeyID),
					testCheckCaptureResourceAttr("apollo_subgraph_api_key.test", "id", &subgraphKeyID),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "name", initialPQLName),
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "variant", fixture.Variant),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "name", initialGraphKeyName),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "role", "GRAPH_ADMIN"),
					resource.TestCheckResourceAttrSet("apollo_graph_api_key.test", "key"),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "variant", fixture.Variant),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "subgraph_name", fixture.Subgraph),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "name", initialSubgraphKeyName),
					resource.TestCheckResourceAttrSet("apollo_subgraph_api_key.test", "key"),
				),
			},
			{
				Config: liveMutableResourcesConfig(
					fixture.GraphID,
					fixture.Variant,
					fixture.Subgraph,
					initialPQLName,
					"Live integration smoke collection updated",
					updatedGraphKeyName,
					"GRAPH_ADMIN",
					updatedSubgraphKeyName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_persisted_query_list.test", "description", "Live integration smoke collection updated"),
					resource.TestCheckResourceAttr("apollo_graph_api_key.test", "name", updatedGraphKeyName),
					resource.TestCheckResourceAttr("apollo_subgraph_api_key.test", "name", updatedSubgraphKeyName),
					resource.TestCheckResourceAttr("apollo_persisted_query_list_link.test", "variant", fixture.Variant),
				),
			},
		},
	})
}

func TestTerraformIntegrationSubgraphResource(t *testing.T) {
	env := requireLiveIntegrationEnv(t, true)
	fixture := bootstrapLiveVariantSubgraphFixture(t, env)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(_ *terraform.State) error {
			checkCtx, checkCancel := liveIntegrationContext(t)
			defer checkCancel()

			_, err := fixture.Client.GetSubgraphMetadata(checkCtx, fixture.GraphID, fixture.Variant, fixture.Subgraph)
			if err == nil {
				return fmt.Errorf("subgraph %s/%s/%s still exists after destroy", fixture.GraphID, fixture.Variant, fixture.Subgraph)
			}
			if !IsErrorKind(err, ErrorKindNotFound) {
				return fmt.Errorf("check subgraph destroy: %w", err)
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:             liveSubgraphResourceConfig(fixture.GraphID, fixture.Variant, fixture.Subgraph),
				ResourceName:       "apollo_subgraph.test",
				ImportState:        true,
				ImportStateId:      JoinImportID(fixture.GraphID, fixture.Variant, fixture.Subgraph),
				ImportStatePersist: true,
			},
			{
				Config: liveSubgraphResourceConfig(fixture.GraphID, fixture.Variant, fixture.Subgraph),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_subgraph.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "variant", fixture.Variant),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "name", fixture.Subgraph),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "routing_url", fixture.RoutingURL),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "revision", fixture.Revision),
					resource.TestCheckResourceAttr("apollo_subgraph.test", "id", JoinImportID(fixture.GraphID, fixture.Variant, fixture.Subgraph)),
				),
			},
		},
	})
}

func TestTerraformIntegrationGraphVariantResource(t *testing.T) {
	env := requireLiveIntegrationEnv(t, true)
	fixture := bootstrapLiveVariantSubgraphFixture(t, env)

	removeCtx, removeCancel := liveIntegrationContext(t)
	err := fixture.Client.RemoveSubgraph(removeCtx, fixture.GraphID, fixture.Variant, fixture.Subgraph)
	removeCancel()
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		t.Fatalf("preparing empty variant fixture failed: %s", err)
	}

	checkCtx, checkCancel := liveIntegrationContext(t)
	variant, err := fixture.Client.GetVariantMetadata(checkCtx, fixture.GraphID, fixture.Variant)
	checkCancel()
	if err != nil {
		t.Fatalf("verify empty variant fixture failed: %s", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(_ *terraform.State) error {
			destroyCtx, destroyCancel := liveIntegrationContext(t)
			defer destroyCancel()

			_, err := fixture.Client.GetVariantMetadata(destroyCtx, fixture.GraphID, fixture.Variant)
			if err == nil {
				return fmt.Errorf("graph variant %s/%s still exists after destroy", fixture.GraphID, fixture.Variant)
			}
			if !IsErrorKind(err, ErrorKindNotFound) {
				return fmt.Errorf("check graph variant destroy: %w", err)
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:             liveGraphVariantResourceConfig(fixture.GraphID, fixture.Variant),
				ResourceName:       "apollo_graph_variant.test",
				ImportState:        true,
				ImportStateId:      JoinImportID(fixture.GraphID, fixture.Variant),
				ImportStatePersist: true,
			},
			{
				Config: liveGraphVariantResourceConfig(fixture.GraphID, fixture.Variant),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "graph_id", fixture.GraphID),
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "variant", fixture.Variant),
					resource.TestCheckResourceAttr("apollo_graph_variant.test", "id", variant.ID),
				),
			},
		},
	})
}
