// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestGraphVariantSubgraphListValue(t *testing.T) {
	t.Parallel()

	list, diags := graphVariantSubgraphListValue([]SubgraphMetadata{
		{
			Name:       "products",
			RoutingURL: " HTTPS://PRODUCTS.EXAMPLE.COM:443/ ",
			Revision:   "rev-products-1",
		},
	})
	if diags.HasError() {
		t.Fatalf("expected helper to succeed, got diagnostics: %#v", diags)
	}

	elements := list.Elements()
	if got, want := len(elements), 1; got != want {
		t.Fatalf("unexpected number of elements: got %d want %d", got, want)
	}

	value, ok := elements[0].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected object value, got %T", elements[0])
	}

	attributes := value.Attributes()
	if got, want := requireStringAttributeValue(t, attributes, "name").ValueString(), "products"; got != want {
		t.Fatalf("unexpected subgraph name: got %q want %q", got, want)
	}
	if got, want := requireStringAttributeValue(t, attributes, "routing_url").ValueString(), "https://products.example.com"; got != want {
		t.Fatalf("unexpected routing URL: got %q want %q", got, want)
	}
	if got, want := requireStringAttributeValue(t, attributes, "revision").ValueString(), "rev-products-1"; got != want {
		t.Fatalf("unexpected revision: got %q want %q", got, want)
	}
}

func TestPersistedQueryListLinkedVariantListValue(t *testing.T) {
	t.Parallel()

	list, diags := persistedQueryListLinkedVariantListValue([]GraphVariantRef{
		{GraphID: " inventory ", Variant: " current "},
	})
	if diags.HasError() {
		t.Fatalf("expected helper to succeed, got diagnostics: %#v", diags)
	}

	elements := list.Elements()
	if got, want := len(elements), 1; got != want {
		t.Fatalf("unexpected number of elements: got %d want %d", got, want)
	}

	value, ok := elements[0].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected object value, got %T", elements[0])
	}

	attributes := value.Attributes()
	if got, want := requireStringAttributeValue(t, attributes, "graph_id").ValueString(), "inventory"; got != want {
		t.Fatalf("unexpected graph ID: got %q want %q", got, want)
	}
	if got, want := requireStringAttributeValue(t, attributes, "variant").ValueString(), "current"; got != want {
		t.Fatalf("unexpected variant: got %q want %q", got, want)
	}
}

func requireStringAttributeValue(t *testing.T, attributes map[string]attr.Value, name string) basetypes.StringValue {
	t.Helper()

	value, ok := attributes[name]
	if !ok {
		t.Fatalf("expected attribute %q to be present", name)
	}

	stringValue, ok := value.(basetypes.StringValue)
	if !ok {
		t.Fatalf("expected attribute %q to be basetypes.StringValue, got %T", name, value)
	}

	return stringValue
}

func TestGraphDataSourceModelFromAPI(t *testing.T) {
	t.Parallel()

	model, diags := graphDataSourceModelFromAPI(context.Background(), "inventory", &GraphMetadata{
		ID:           "inventory",
		Name:         "Inventory Graph",
		Title:        "Inventory",
		VariantNames: []string{"current", "staging"},
	})
	if diags.HasError() {
		t.Fatalf("expected helper to succeed, got diagnostics: %#v", diags)
	}

	if got, want := model.Name.ValueString(), "Inventory Graph"; got != want {
		t.Fatalf("unexpected graph name: got %q want %q", got, want)
	}
	if got, want := len(model.VariantNames.Elements()), 2; got != want {
		t.Fatalf("unexpected variant count: got %d want %d", got, want)
	}
}

func TestSubgraphDataSourceModelFromAPI(t *testing.T) {
	t.Parallel()

	model := subgraphDataSourceModelFromAPI(&SubgraphMetadata{
		GraphID:    "inventory",
		Variant:    "current",
		Name:       "products",
		RoutingURL: "https://products.example.com/graphql",
		Revision:   "rev-products-1",
	})

	if got, want := model.ID.ValueString(), "inventory:current:products"; got != want {
		t.Fatalf("unexpected synthetic ID: got %q want %q", got, want)
	}
	if got, want := model.Revision.ValueString(), "rev-products-1"; got != want {
		t.Fatalf("unexpected revision: got %q want %q", got, want)
	}
}
