// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	graphReadScope              = "Graph read or broader org access"
	graphVariantReadScope       = "Graph read or broader org access for variant metadata"
	subgraphReadScope           = "Graph read or broader org access for subgraph metadata"
	persistedQueryListReadScope = "Graph or org access for persisted query list management"
)

var (
	graphVariantSubgraphAttrTypes = map[string]attr.Type{
		"name":        types.StringType,
		"routing_url": types.StringType,
		"revision":    types.StringType,
	}
	graphVariantSubgraphObjectType = types.ObjectType{AttrTypes: graphVariantSubgraphAttrTypes}

	persistedQueryListLinkedVariantAttrTypes = map[string]attr.Type{
		"graph_id": types.StringType,
		"variant":  types.StringType,
	}
	persistedQueryListLinkedVariantObjectType = types.ObjectType{AttrTypes: persistedQueryListLinkedVariantAttrTypes}
)

func nullableStringValue(value string) types.String {
	if normalized := NormalizeString(value); normalized != "" {
		return types.StringValue(normalized)
	}

	return types.StringNull()
}

func stringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	normalized := make([]string, len(values))
	for index, value := range values {
		normalized[index] = NormalizeString(value)
	}

	return types.ListValueFrom(ctx, types.StringType, normalized)
}

func graphVariantSubgraphListValue(subgraphs []SubgraphMetadata) (types.List, diag.Diagnostics) {
	values := make([]attr.Value, 0, len(subgraphs))
	var diags diag.Diagnostics

	for _, subgraph := range subgraphs {
		value, valueDiags := types.ObjectValue(graphVariantSubgraphAttrTypes, map[string]attr.Value{
			"name":        nullableStringValue(subgraph.Name),
			"routing_url": nullableStringValue(NormalizeURLString(subgraph.RoutingURL)),
			"revision":    nullableStringValue(subgraph.Revision),
		})
		diags.Append(valueDiags...)
		values = append(values, value)
	}

	listValue, listDiags := types.ListValue(graphVariantSubgraphObjectType, values)
	diags.Append(listDiags...)
	return listValue, diags
}

func persistedQueryListLinkedVariantListValue(linkedVariants []GraphVariantRef) (types.List, diag.Diagnostics) {
	values := make([]attr.Value, 0, len(linkedVariants))
	var diags diag.Diagnostics

	for _, linkedVariant := range linkedVariants {
		value, valueDiags := types.ObjectValue(persistedQueryListLinkedVariantAttrTypes, map[string]attr.Value{
			"graph_id": nullableStringValue(linkedVariant.GraphID),
			"variant":  nullableStringValue(linkedVariant.Variant),
		})
		diags.Append(valueDiags...)
		values = append(values, value)
	}

	listValue, listDiags := types.ListValue(persistedQueryListLinkedVariantObjectType, values)
	diags.Append(listDiags...)
	return listValue, diags
}
