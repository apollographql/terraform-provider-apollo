// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientIntegrationReadOnly(t *testing.T) {
	if os.Getenv("APOLLO_PLATFORM_INTEGRATION") != "1" {
		t.Skip("set APOLLO_PLATFORM_INTEGRATION=1 to run GraphOS integration tests")
	}

	apiKey := os.Getenv(apiKeyEnvVar)
	graphID := os.Getenv("APOLLO_TEST_GRAPH_ID")
	if apiKey == "" || graphID == "" {
		t.Skip("set APOLLO_KEY and APOLLO_TEST_GRAPH_ID to run read-only GraphOS integration tests")
	}

	endpoint := os.Getenv(endpointEnvVar)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	client := NewClient(apiKey, endpoint, "terraform-provider-apollo/integration")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := client.GetGraphMetadata(ctx, graphID); err != nil {
		t.Fatalf("graph metadata lookup failed: %s", err)
	}

	variant := os.Getenv("APOLLO_TEST_VARIANT")
	if variant != "" {
		if _, err := client.GetVariantMetadata(ctx, graphID, variant); err != nil {
			t.Fatalf("variant metadata lookup failed: %s", err)
		}

		if _, err := client.GetVariantPersistedQueryList(ctx, graphID, variant); err != nil && !IsErrorKind(err, ErrorKindNotFound) {
			t.Fatalf("variant persisted query list lookup failed: %s", err)
		}
	}

	subgraph := os.Getenv("APOLLO_TEST_SUBGRAPH")
	if variant != "" && subgraph != "" {
		if _, err := client.GetSubgraphMetadata(ctx, graphID, variant, subgraph); err != nil {
			t.Fatalf("subgraph metadata lookup failed: %s", err)
		}
	}
}

func TestClientIntegrationPersistedQueryListSmoke(t *testing.T) {
	if os.Getenv("APOLLO_PLATFORM_INTEGRATION") != "1" || os.Getenv("APOLLO_INTEGRATION_ALLOW_MUTATIONS") != "1" {
		t.Skip("set APOLLO_PLATFORM_INTEGRATION=1 and APOLLO_INTEGRATION_ALLOW_MUTATIONS=1 to run mutating GraphOS integration tests")
	}

	apiKey := os.Getenv(apiKeyEnvVar)
	graphID := os.Getenv("APOLLO_TEST_GRAPH_ID")
	if apiKey == "" || graphID == "" {
		t.Skip("set APOLLO_KEY and APOLLO_TEST_GRAPH_ID to run the PQL smoke test")
	}

	endpoint := os.Getenv(endpointEnvVar)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	client := NewClient(apiKey, endpoint, "terraform-provider-apollo/integration")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name := fmt.Sprintf("terraform-provider-apollo-%d", time.Now().UnixNano())
	pql, err := client.CreatePersistedQueryList(ctx, PersistedQueryListInput{
		GraphID:     graphID,
		Name:        name,
		Description: "Disposable integration test collection",
	})
	if err != nil {
		t.Fatalf("create persisted query list failed: %s", err)
	}

	variant := os.Getenv("APOLLO_TEST_VARIANT")
	if variant != "" {
		if _, err := client.LinkPersistedQueryListToVariant(ctx, PersistedQueryListLinkInput{
			PersistedQueryListID: pql.ID,
			GraphID:              graphID,
			Variant:              variant,
		}); err != nil {
			t.Fatalf("link persisted query list failed: %s", err)
		}

		if err := client.UnlinkPersistedQueryListFromVariant(ctx, PersistedQueryListLinkInput{
			PersistedQueryListID: pql.ID,
			GraphID:              graphID,
			Variant:              variant,
		}); err != nil {
			t.Fatalf("unlink persisted query list failed: %s", err)
		}
	}

	if err := client.DeletePersistedQueryList(ctx, graphID, pql.ID); err != nil {
		t.Fatalf("delete persisted query list failed: %s", err)
	}
}
