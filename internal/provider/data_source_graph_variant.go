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

var _ datasource.DataSource = &graphVariantDataSource{}

type graphVariantDataSource struct {
	client *Client
}

type graphVariantDataSourceModel struct {
	GraphID   types.String `tfsdk:"graph_id"`
	Variant   types.String `tfsdk:"variant"`
	ID        types.String `tfsdk:"id"`
	Subgraphs types.List   `tfsdk:"subgraphs"`
}

func NewGraphVariantDataSource() datasource.DataSource {
	return &graphVariantDataSource{}
}

func (d *graphVariantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_variant"
}

func (d *graphVariantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Reads GraphOS variant metadata, including the visible subgraph topology for that variant.",
		Attributes: map[string]dschema.Attribute{
			"graph_id": dschema.StringAttribute{
				MarkdownDescription: "Graph identifier to look up.",
				Required:            true,
			},
			"variant": dschema.StringAttribute{
				MarkdownDescription: "Variant name to look up.",
				Required:            true,
			},
			"id": dschema.StringAttribute{
				MarkdownDescription: "Stable graph variant identifier returned by GraphOS.",
				Computed:            true,
			},
			"subgraphs": dschema.ListNestedAttribute{
				MarkdownDescription: "Subgraphs currently visible on the variant.",
				Computed:            true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"name": dschema.StringAttribute{
							MarkdownDescription: "Subgraph name.",
							Computed:            true,
						},
						"routing_url": dschema.StringAttribute{
							MarkdownDescription: "Routing URL registered for the subgraph.",
							Computed:            true,
						},
						"revision": dschema.StringAttribute{
							MarkdownDescription: "Current GraphOS revision identifier for the subgraph registration, when available.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *graphVariantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics, "data source")
}

func (d *graphVariantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config graphVariantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variant, err := d.client.GetVariantMetadata(ctx, config.GraphID.ValueString(), config.Variant.ValueString())
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read graph variant", err, graphVariantReadScope)
		return
	}

	state, stateDiags := graphVariantDataSourceModelFromAPI(variant)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func graphVariantDataSourceModelFromAPI(variant *VariantMetadata) (graphVariantDataSourceModel, diag.Diagnostics) {
	subgraphs, diags := graphVariantSubgraphListValue(variant.Subgraphs)

	return graphVariantDataSourceModel{
		GraphID:   types.StringValue(NormalizeString(variant.GraphID)),
		Variant:   nullableStringValue(variant.Name),
		ID:        nullableStringValue(variant.ID),
		Subgraphs: subgraphs,
	}, diags
}
