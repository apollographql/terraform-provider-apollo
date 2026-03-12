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

var _ datasource.DataSource = &persistedQueryListDataSource{}

type persistedQueryListDataSource struct {
	client *Client
}

type persistedQueryListDataSourceModel struct {
	GraphID        types.String `tfsdk:"graph_id"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	LinkedVariants types.List   `tfsdk:"linked_variants"`
}

func NewPersistedQueryListDataSource() datasource.DataSource {
	return &persistedQueryListDataSource{}
}

func (d *persistedQueryListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_persisted_query_list"
}

func (d *persistedQueryListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Reads one GraphOS persisted query list by graph identifier and persisted query list identifier.",
		Attributes: map[string]dschema.Attribute{
			"graph_id": dschema.StringAttribute{
				MarkdownDescription: "Graph identifier that owns the persisted query list.",
				Required:            true,
			},
			"id": dschema.StringAttribute{
				MarkdownDescription: "Persisted query list identifier to look up.",
				Required:            true,
			},
			"name": dschema.StringAttribute{
				MarkdownDescription: "Persisted query list name returned by GraphOS.",
				Computed:            true,
			},
			"linked_variants": dschema.ListNestedAttribute{
				MarkdownDescription: "Variants currently linked to the persisted query list.",
				Computed:            true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"graph_id": dschema.StringAttribute{
							MarkdownDescription: "Graph identifier of the linked variant.",
							Computed:            true,
						},
						"variant": dschema.StringAttribute{
							MarkdownDescription: "Variant name linked to the persisted query list.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *persistedQueryListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics, "data source")
}

func (d *persistedQueryListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config persistedQueryListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pql, err := d.client.GetPersistedQueryList(ctx, config.GraphID.ValueString(), config.ID.ValueString())
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read persisted query list", err, persistedQueryListReadScope)
		return
	}

	state, stateDiags := persistedQueryListDataSourceModelFromAPI(config.GraphID.ValueString(), pql)
	resp.Diagnostics.Append(stateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func persistedQueryListDataSourceModelFromAPI(graphID string, pql *PersistedQueryList) (persistedQueryListDataSourceModel, diag.Diagnostics) {
	linkedVariants, diags := persistedQueryListLinkedVariantListValue(pql.LinkedVariants)

	return persistedQueryListDataSourceModel{
		GraphID:        types.StringValue(NormalizeString(graphID)),
		ID:             nullableStringValue(pql.ID),
		Name:           nullableStringValue(pql.Name),
		LinkedVariants: linkedVariants,
	}, diags
}
