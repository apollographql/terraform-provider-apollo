// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateGraphAPIKeyRoleValue(t *testing.T) {
	t.Parallel()

	if err := validateGraphAPIKeyRoleValue(" GRAPH_ADMIN "); err != nil {
		t.Fatalf("expected GRAPH_ADMIN to validate, got %s", err)
	}

	if err := validateGraphAPIKeyRoleValue("NOT_A_ROLE"); err == nil {
		t.Fatal("expected invalid role to fail validation")
	}
}

func TestPreserveAPIKeyToken(t *testing.T) {
	t.Parallel()

	if got := preserveAPIKeyToken(types.StringValue("service:old"), ""); got.ValueString() != "service:old" {
		t.Fatalf("expected prior token to be preserved, got %q", got.ValueString())
	}

	if got := preserveAPIKeyToken(types.StringNull(), " service:new "); got.ValueString() != "service:new" {
		t.Fatalf("expected new token to normalize, got %q", got.ValueString())
	}

	if got := preserveAPIKeyToken(types.StringNull(), ""); !got.IsNull() {
		t.Fatalf("expected missing token without prior state to stay null, got %#v", got)
	}
}

func TestGraphAPIKeyImportStateInvalidID(t *testing.T) {
	t.Parallel()

	resourceUnderTest := &graphAPIKeyResource{}
	var resp tfresource.ImportStateResponse

	resourceUnderTest.ImportState(context.Background(), tfresource.ImportStateRequest{ID: "inventory"}, &resp)

	if !diagnosticsContainSummary(resp.Diagnostics, "Import graph API key") {
		t.Fatalf("expected import diagnostic, got %#v", resp.Diagnostics)
	}
}

func TestSubgraphAPIKeyImportStateInvalidID(t *testing.T) {
	t.Parallel()

	resourceUnderTest := &subgraphAPIKeyResource{}
	var resp tfresource.ImportStateResponse

	resourceUnderTest.ImportState(context.Background(), tfresource.ImportStateRequest{ID: "inventory:production:key-1"}, &resp)

	if !diagnosticsContainSummary(resp.Diagnostics, "Import subgraph API key") {
		t.Fatalf("expected import diagnostic, got %#v", resp.Diagnostics)
	}
}

func TestGraphAPIKeyResourceModelFromAPIPreservesToken(t *testing.T) {
	t.Parallel()

	model := graphAPIKeyResourceModelFromAPI(&GraphAPIKey{
		ID:         " graph-key-1 ",
		GraphID:    " inventory ",
		Name:       " deploy ",
		Role:       " GRAPH_ADMIN ",
		CreatedAt:  " 2026-03-11T00:00:00Z ",
		LastUsedAt: "",
	}, types.StringValue("service:preserved"))

	if got, want := model.ID.ValueString(), "graph-key-1"; got != want {
		t.Fatalf("unexpected id: got %q want %q", got, want)
	}

	if got, want := model.Key.ValueString(), "service:preserved"; got != want {
		t.Fatalf("expected token preservation, got %q want %q", got, want)
	}
}

func TestSubgraphAPIKeyResourceModelFromAPIPreservesToken(t *testing.T) {
	t.Parallel()

	model := subgraphAPIKeyResourceModelFromAPI(&SubgraphAPIKey{
		ID:           " subgraph-key-1 ",
		GraphID:      " inventory ",
		Variant:      " production ",
		SubgraphName: " products ",
		Name:         " products-router ",
		CreatedAt:    " 2026-03-11T00:00:00Z ",
		LastUsedAt:   "",
	}, types.StringValue("service:preserved"))

	if got, want := model.ID.ValueString(), "subgraph-key-1"; got != want {
		t.Fatalf("unexpected id: got %q want %q", got, want)
	}

	if got, want := model.Key.ValueString(), "service:preserved"; got != want {
		t.Fatalf("expected token preservation, got %q want %q", got, want)
	}
}
