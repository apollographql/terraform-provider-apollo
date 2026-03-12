// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	graphVariantReadResourceScope   = "Graph read or broader org access for variant metadata"
	graphVariantDeleteResourceScope = "Graph admin or broader org access for variant deletion"
)

var _ resource.Resource = &graphVariantResource{}
var _ resource.ResourceWithImportState = &graphVariantResource{}

type graphVariantResource struct {
	client *Client
}

type graphVariantResourceModel struct {
	ID      types.String `tfsdk:"id"`
	GraphID types.String `tfsdk:"graph_id"`
	Variant types.String `tfsdk:"variant"`
}

func NewGraphVariantResource() resource.Resource {
	return &graphVariantResource{}
}

func (r *graphVariantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_variant"
}

func (r *graphVariantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Represents an existing GraphOS graph variant. New variants are created in GraphOS by publishing the first subgraph, so this resource supports brownfield import, read, and delete but does not bootstrap new variants.",
		Attributes: map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Graph identifier, for example `inventory`.",
				Required:            true,
				PlanModifiers:       graphVariantReplaceModifiers(),
			},
			"variant": rschema.StringAttribute{
				MarkdownDescription: "Variant name, for example `production`.",
				Required:            true,
				PlanModifiers:       graphVariantReplaceModifiers(),
			},
			"id": scaffoldResourceIDAttribute("Remote GraphOS variant identifier. Import format: `graph_id:variant`."),
		},
	}
}

func (r *graphVariantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *graphVariantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan graphVariantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	AddAPIErrorDiagnostic(
		&resp.Diagnostics,
		"Create graph variant",
		capabilityDeferredError(
			operationCreateGraphVariant,
			"GraphOS creates variants by publishing the first subgraph. `apollo_graph_variant` does not model that bootstrap workflow, so create is deferred. Import an existing variant with `graph_id:variant` instead.",
		),
		"",
	)
}

func (r *graphVariantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state graphVariantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variant, err := r.client.GetVariantMetadata(ctx, state.GraphID.ValueString(), state.Variant.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read graph variant", err, graphVariantReadResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, graphVariantResourceModelFromAPI(variant))...)
}

func (r *graphVariantResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update graph variant: provider bug",
		"`apollo_graph_variant` is replacement-only in this provider slice. Terraform should have planned a replace instead of calling Update.",
	)
}

func (r *graphVariantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state graphVariantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGraphVariant(ctx, state.GraphID.ValueString(), state.Variant.ValueString())
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete graph variant", err, graphVariantDeleteResourceScope)
	}
}

func (r *graphVariantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Import graph variant", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variant"), parts[1])...)
}

func graphVariantResourceModelFromAPI(variant *VariantMetadata) graphVariantResourceModel {
	return graphVariantResourceModel{
		ID:      types.StringValue(NormalizeString(variant.ID)),
		GraphID: types.StringValue(NormalizeString(variant.GraphID)),
		Variant: types.StringValue(NormalizeString(variant.Name)),
	}
}

func graphVariantReplaceModifiers() []planmodifier.String {
	return []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
}
