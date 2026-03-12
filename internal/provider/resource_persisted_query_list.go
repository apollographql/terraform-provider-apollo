// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const persistedQueryListManagementScope = "Graph or org access for persisted query list management"

var _ resource.Resource = &persistedQueryListResource{}
var _ resource.ResourceWithImportState = &persistedQueryListResource{}

type persistedQueryListResource struct {
	client *Client
}

type persistedQueryListResourceModel struct {
	ID          types.String `tfsdk:"id"`
	GraphID     types.String `tfsdk:"graph_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewPersistedQueryListResource() resource.Resource {
	return &persistedQueryListResource{}
}

func (r *persistedQueryListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_persisted_query_list"
}

func (r *persistedQueryListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages a GraphOS persisted query list within a specific graph. The current Platform API allows writing `description` during create and update, but does not expose that field for readback, so imported resources leave `description` unset until the next apply.",
		Attributes: map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Graph identifier that owns the persisted query list.",
				Required:            true,
			},
			"name": rschema.StringAttribute{
				MarkdownDescription: "Persisted query list name.",
				Required:            true,
			},
			"description": rschema.StringAttribute{
				MarkdownDescription: "Optional description sent to GraphOS on create and update. The current public schema does not return it on read.",
				Optional:            true,
			},
			"id": scaffoldResourceIDAttribute("Resource identifier. Import format: `graph_id:persisted_query_list_id`."),
		},
	}
}

func (r *persistedQueryListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *persistedQueryListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan persistedQueryListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePersistedQueryList(ctx, PersistedQueryListInput{
		GraphID:     NormalizeString(plan.GraphID.ValueString()),
		Name:        NormalizeString(plan.Name.ValueString()),
		Description: NormalizeString(plan.Description.ValueString()),
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Create persisted query list", err, persistedQueryListManagementScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, persistedQueryListModelFromAPI(plan.GraphID.ValueString(), created, plan.Description.ValueString()))...)
}

func (r *persistedQueryListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state persistedQueryListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pql, err := r.client.GetPersistedQueryList(ctx, state.GraphID.ValueString(), state.ID.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read persisted query list", err, persistedQueryListManagementScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, persistedQueryListModelFromAPI(state.GraphID.ValueString(), pql, state.Description.ValueString()))...)
}

func (r *persistedQueryListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan persistedQueryListResourceModel
	var state persistedQueryListResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdatePersistedQueryList(ctx, PersistedQueryListInput{
		GraphID:     NormalizeString(state.GraphID.ValueString()),
		ID:          state.ID.ValueString(),
		Name:        NormalizeString(plan.Name.ValueString()),
		Description: NormalizeString(plan.Description.ValueString()),
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Update persisted query list", err, persistedQueryListManagementScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, persistedQueryListModelFromAPI(state.GraphID.ValueString(), updated, plan.Description.ValueString()))...)
}

func (r *persistedQueryListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state persistedQueryListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePersistedQueryList(ctx, state.GraphID.ValueString(), state.ID.ValueString())
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete persisted query list", err, persistedQueryListManagementScope)
	}
}

func (r *persistedQueryListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Import persisted query list", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func persistedQueryListModelFromAPI(graphID string, pql *PersistedQueryList, description string) persistedQueryListResourceModel {
	model := persistedQueryListResourceModel{
		ID:      types.StringValue(NormalizeString(pql.ID)),
		GraphID: types.StringValue(NormalizeString(graphID)),
		Name:    types.StringValue(NormalizeString(pql.Name)),
	}

	if normalizedDescription := NormalizeString(description); normalizedDescription == "" {
		model.Description = types.StringNull()
	} else {
		model.Description = types.StringValue(normalizedDescription)
	}

	return model
}
