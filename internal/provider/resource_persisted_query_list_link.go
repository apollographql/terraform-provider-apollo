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

const persistedQueryListLinkScope = "Graph admin or broader org access for persisted query list variant linkage"

var _ resource.Resource = &persistedQueryListLinkResource{}
var _ resource.ResourceWithImportState = &persistedQueryListLinkResource{}

type persistedQueryListLinkResource struct {
	client *Client
}

type persistedQueryListLinkResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	PersistedQueryListID types.String `tfsdk:"persisted_query_list_id"`
	GraphID              types.String `tfsdk:"graph_id"`
	Variant              types.String `tfsdk:"variant"`
}

func NewPersistedQueryListLinkResource() resource.Resource {
	return &persistedQueryListLinkResource{}
}

func (r *persistedQueryListLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_persisted_query_list_link"
}

func (r *persistedQueryListLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Represents the association between a persisted query list and a GraphOS variant. This resource is immutable in place because GraphOS models the link as a single variant-scoped attachment.",
		Attributes: map[string]rschema.Attribute{
			"persisted_query_list_id": rschema.StringAttribute{
				MarkdownDescription: "Identifier of the persisted query list to associate.",
				Required:            true,
				PlanModifiers:       persistedQueryListLinkReplaceModifiers(),
			},
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Graph identifier that owns the target variant.",
				Required:            true,
				PlanModifiers:       persistedQueryListLinkReplaceModifiers(),
			},
			"variant": rschema.StringAttribute{
				MarkdownDescription: "Variant that should receive the persisted query list association.",
				Required:            true,
				PlanModifiers:       persistedQueryListLinkReplaceModifiers(),
			},
			"id": scaffoldResourceIDAttribute("Resource identifier. Import format: `persisted_query_list_id:graph_id:variant`."),
		},
	}
}

func (r *persistedQueryListLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *persistedQueryListLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan persistedQueryListLinkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := persistedQueryListLinkInputFromModel(plan)
	linked, err := r.client.LinkPersistedQueryListToVariant(ctx, input)
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Create persisted query list link", err, persistedQueryListLinkScope)
		return
	}

	if linked == nil || linked.ID != input.PersistedQueryListID || !persistedQueryListHasVariantLink(linked, input.GraphID, input.Variant) {
		resp.Diagnostics.AddError(
			"Create persisted query list link: malformed GraphOS response",
			"The GraphOS Platform API did not confirm the requested persisted query list association after the link mutation completed.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, persistedQueryListLinkModelFromInput(input))...)
}

func (r *persistedQueryListLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state persistedQueryListLinkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetVariantPersistedQueryList(ctx, state.GraphID.ValueString(), state.Variant.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read persisted query list link", err, persistedQueryListLinkScope)
		return
	}

	if current == nil || current.ID != NormalizeString(state.PersistedQueryListID.ValueString()) || !persistedQueryListHasVariantLink(current, state.GraphID.ValueString(), state.Variant.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, persistedQueryListLinkModelFromInput(persistedQueryListLinkInputFromModel(state)))...)
}

func (r *persistedQueryListLinkResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update persisted query list link: provider bug",
		"`apollo_persisted_query_list_link` is replacement-only. Terraform should have planned a replace instead of calling Update.",
	)
}

func (r *persistedQueryListLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state persistedQueryListLinkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetVariantPersistedQueryList(ctx, state.GraphID.ValueString(), state.Variant.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete persisted query list link", err, persistedQueryListLinkScope)
		return
	}

	if current == nil || current.ID != NormalizeString(state.PersistedQueryListID.ValueString()) {
		return
	}

	if err := r.client.UnlinkPersistedQueryListFromVariant(ctx, persistedQueryListLinkInputFromModel(state)); err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete persisted query list link", err, persistedQueryListLinkScope)
	}
}

func (r *persistedQueryListLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 3)
	if err != nil {
		resp.Diagnostics.AddError("Import persisted query list link", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("persisted_query_list_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variant"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), JoinImportID(parts[0], parts[1], parts[2]))...)
}

func persistedQueryListLinkReplaceModifiers() []planmodifier.String {
	return []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
}

func persistedQueryListLinkInputFromModel(model persistedQueryListLinkResourceModel) PersistedQueryListLinkInput {
	return PersistedQueryListLinkInput{
		PersistedQueryListID: NormalizeString(model.PersistedQueryListID.ValueString()),
		GraphID:              NormalizeString(model.GraphID.ValueString()),
		Variant:              NormalizeString(model.Variant.ValueString()),
	}
}

func persistedQueryListLinkModelFromInput(input PersistedQueryListLinkInput) persistedQueryListLinkResourceModel {
	return persistedQueryListLinkResourceModel{
		ID:                   types.StringValue(JoinImportID(input.PersistedQueryListID, input.GraphID, input.Variant)),
		PersistedQueryListID: types.StringValue(NormalizeString(input.PersistedQueryListID)),
		GraphID:              types.StringValue(NormalizeString(input.GraphID)),
		Variant:              types.StringValue(NormalizeString(input.Variant)),
	}
}
