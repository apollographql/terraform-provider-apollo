// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &graphDataSource{}

type graphDataSource struct {
	client *Client
}

type graphDataSourceModel struct {
	GraphID      types.String `tfsdk:"graph_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Title        types.String `tfsdk:"title"`
	VariantNames types.List   `tfsdk:"variant_names"`
}

func NewGraphDataSource() datasource.DataSource {
	return &graphDataSource{}
}

func (d *graphDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph"
}

func (d *graphDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Reads GraphOS graph metadata for brownfield discovery and module composition.",
		Attributes: map[string]dschema.Attribute{
			"graph_id": dschema.StringAttribute{
				MarkdownDescription: "Graph identifier to look up.",
				Required:            true,
			},
			"id": dschema.StringAttribute{
				MarkdownDescription: "Stable graph identifier returned by GraphOS.",
				Computed:            true,
			},
			"name": dschema.StringAttribute{
				MarkdownDescription: "Graph name returned by GraphOS.",
				Computed:            true,
			},
			"title": dschema.StringAttribute{
				MarkdownDescription: "Optional graph title returned by GraphOS.",
				Computed:            true,
			},
			"variant_names": dschema.ListAttribute{
				MarkdownDescription: "Variant names currently visible on the graph.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *graphDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics, "data source")
}

func (d *graphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config graphDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	graph, err := d.client.GetGraphMetadata(ctx, config.GraphID.ValueString())
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read graph", err, graphReadScope)
		return
	}

	state, stateDiags := graphDataSourceModelFromAPI(ctx, config.GraphID.ValueString(), graph)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func graphDataSourceModelFromAPI(ctx context.Context, graphID string, graph *GraphMetadata) (graphDataSourceModel, diag.Diagnostics) {
	variantNames, diags := stringListValue(ctx, graph.VariantNames)

	return graphDataSourceModel{
		GraphID:      types.StringValue(NormalizeString(graphID)),
		ID:           nullableStringValue(graph.ID),
		Name:         nullableStringValue(graph.Name),
		Title:        nullableStringValue(graph.Title),
		VariantNames: variantNames,
	}, diags
}
