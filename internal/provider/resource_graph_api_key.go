// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &graphAPIKeyResource{}
var _ resource.ResourceWithImportState = &graphAPIKeyResource{}

type graphAPIKeyResource struct {
	client *Client
}

type graphAPIKeyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	GraphID    types.String `tfsdk:"graph_id"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Key        types.String `tfsdk:"key"`
	CreatedAt  types.String `tfsdk:"created_at"`
	LastUsedAt types.String `tfsdk:"last_used_at"`
}

func NewGraphAPIKeyResource() resource.Resource {
	return &graphAPIKeyResource{}
}

func (r *graphAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_api_key"
}

func (r *graphAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages a GraphOS graph-scoped API key. The `key` value is only visible when GraphOS creates the key, so this resource stores that sensitive value in Terraform state for Terraform-created resources and intentionally cannot rehydrate it on import or refresh.",
		Attributes: map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Graph identifier that owns the API key.",
				Required:            true,
				PlanModifiers:       apiKeyReplacePlanModifiers(),
			},
			"name": rschema.StringAttribute{
				MarkdownDescription: "Display name for the API key.",
				Required:            true,
			},
			"role": rschema.StringAttribute{
				MarkdownDescription: "GraphOS role to assign when creating the key. GraphOS does not allow changing roles after creation, so Terraform replaces the resource when this value changes.",
				Required:            true,
				Validators:          []validator.String{graphAPIKeyRoleValidator()},
				PlanModifiers:       apiKeyReplacePlanModifiers(),
			},
			"key":          apiKeyComputedSensitiveStringAttribute("Sensitive API key value returned by GraphOS only at creation time. Imported resources and refreshes do not rehydrate this field."),
			"created_at":   apiKeyComputedStringAttribute("Timestamp when GraphOS created the API key."),
			"last_used_at": apiKeyComputedStringAttribute("Timestamp when GraphOS last observed this API key being used, when available."),
			"id":           scaffoldResourceIDAttribute("Remote GraphOS API key identifier. Import format: `graph_id:key_id`."),
		},
	}
}

func (r *graphAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *graphAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan graphAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := NormalizeString(plan.Role.ValueString())
	if err := validateGraphAPIKeyRoleValue(role); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("role"), "Invalid graph API key role", err.Error())
		return
	}

	key, err := r.client.CreateGraphAPIKey(ctx, GraphAPIKeyInput{
		GraphID: NormalizeString(plan.GraphID.ValueString()),
		Name:    NormalizeString(plan.Name.ValueString()),
		Role:    role,
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Create graph API key", err, graphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, graphAPIKeyResourceModelFromAPI(key, types.StringNull()))...)
}

func (r *graphAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state graphAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetGraphAPIKey(ctx, state.GraphID.ValueString(), state.ID.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read graph API key", err, graphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, graphAPIKeyResourceModelFromAPI(key, state.Key))...)
}

func (r *graphAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan graphAPIKeyResourceModel
	var state graphAPIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.UpdateGraphAPIKey(ctx, GraphAPIKeyInput{
		GraphID: NormalizeString(state.GraphID.ValueString()),
		ID:      NormalizeString(state.ID.ValueString()),
		Name:    NormalizeString(plan.Name.ValueString()),
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Update graph API key", err, graphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, graphAPIKeyResourceModelFromAPI(key, state.Key))...)
}

func (r *graphAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state graphAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetGraphAPIKey(ctx, state.GraphID.ValueString(), state.ID.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete graph API key", err, graphAPIKeyResourceScope)
		return
	}

	err = r.client.DeleteGraphAPIKey(ctx, state.GraphID.ValueString(), state.ID.ValueString())
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete graph API key", err, graphAPIKeyResourceScope)
	}
}

func (r *graphAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Import graph API key", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func graphAPIKeyResourceModelFromAPI(key *GraphAPIKey, previousToken types.String) graphAPIKeyResourceModel {
	return graphAPIKeyResourceModel{
		ID:         types.StringValue(NormalizeString(key.ID)),
		GraphID:    types.StringValue(NormalizeString(key.GraphID)),
		Name:       types.StringValue(NormalizeString(key.Name)),
		Role:       types.StringValue(NormalizeString(key.Role)),
		Key:        preserveAPIKeyToken(previousToken, key.Token),
		CreatedAt:  nullableStringValue(key.CreatedAt),
		LastUsedAt: nullableStringValue(key.LastUsedAt),
	}
}
