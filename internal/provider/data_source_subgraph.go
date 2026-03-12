// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &subgraphDataSource{}

type subgraphDataSource struct {
	client *Client
}

type subgraphDataSourceModel struct {
	GraphID    types.String `tfsdk:"graph_id"`
	Variant    types.String `tfsdk:"variant"`
	Name       types.String `tfsdk:"name"`
	ID         types.String `tfsdk:"id"`
	RoutingURL types.String `tfsdk:"routing_url"`
	Revision   types.String `tfsdk:"revision"`
}

func NewSubgraphDataSource() datasource.DataSource {
	return &subgraphDataSource{}
}

func (d *subgraphDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph"
}

func (d *subgraphDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Reads GraphOS subgraph registration metadata for a specific graph variant.",
		Attributes: map[string]dschema.Attribute{
			"graph_id": dschema.StringAttribute{
				MarkdownDescription: "Owning graph identifier.",
				Required:            true,
			},
			"variant": dschema.StringAttribute{
				MarkdownDescription: "Variant that owns the subgraph registration.",
				Required:            true,
			},
			"name": dschema.StringAttribute{
				MarkdownDescription: "Subgraph name to look up.",
				Required:            true,
			},
			"id": dschema.StringAttribute{
				MarkdownDescription: "Synthetic stable identifier in the form `graph_id:variant:name`.",
				Computed:            true,
			},
			"routing_url": dschema.StringAttribute{
				MarkdownDescription: "Routing URL returned by GraphOS.",
				Computed:            true,
			},
			"revision": dschema.StringAttribute{
				MarkdownDescription: "Current GraphOS revision identifier for the registration, when available.",
				Computed:            true,
			},
		},
	}
}

func (d *subgraphDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics, "data source")
}

func (d *subgraphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subgraphDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subgraph, err := d.client.GetSubgraphMetadata(ctx, config.GraphID.ValueString(), config.Variant.ValueString(), config.Name.ValueString())
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read subgraph", err, subgraphReadScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, subgraphDataSourceModelFromAPI(subgraph))...)
}

func subgraphDataSourceModelFromAPI(subgraph *SubgraphMetadata) subgraphDataSourceModel {
	return subgraphDataSourceModel{
		GraphID:    nullableStringValue(subgraph.GraphID),
		Variant:    nullableStringValue(subgraph.Variant),
		Name:       nullableStringValue(subgraph.Name),
		ID:         types.StringValue(JoinImportID(subgraph.GraphID, subgraph.Variant, subgraph.Name)),
		RoutingURL: nullableStringValue(NormalizeURLString(subgraph.RoutingURL)),
		Revision:   nullableStringValue(subgraph.Revision),
	}
}
