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

var _ resource.Resource = &subgraphAPIKeyResource{}
var _ resource.ResourceWithImportState = &subgraphAPIKeyResource{}

type subgraphAPIKeyResource struct {
	client *Client
}

type subgraphAPIKeyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	GraphID      types.String `tfsdk:"graph_id"`
	Variant      types.String `tfsdk:"variant"`
	SubgraphName types.String `tfsdk:"subgraph_name"`
	Name         types.String `tfsdk:"name"`
	Key          types.String `tfsdk:"key"`
	CreatedAt    types.String `tfsdk:"created_at"`
	LastUsedAt   types.String `tfsdk:"last_used_at"`
}

func NewSubgraphAPIKeyResource() resource.Resource {
	return &subgraphAPIKeyResource{}
}

func (r *subgraphAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph_api_key"
}

func (r *subgraphAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages a single-target GraphOS subgraph API key. The `key` value is only visible at create time, so this resource preserves it in sensitive Terraform state for Terraform-created resources and intentionally cannot rehydrate it on import or refresh.",
		Attributes: map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Owning graph identifier for the targeted subgraph registration.",
				Required:            true,
				PlanModifiers:       apiKeyReplacePlanModifiers(),
			},
			"variant": rschema.StringAttribute{
				MarkdownDescription: "Variant that owns the targeted subgraph registration.",
				Required:            true,
				PlanModifiers:       apiKeyReplacePlanModifiers(),
			},
			"subgraph_name": rschema.StringAttribute{
				MarkdownDescription: "Subgraph that receives the API key.",
				Required:            true,
				PlanModifiers:       apiKeyReplacePlanModifiers(),
			},
			"name": rschema.StringAttribute{
				MarkdownDescription: "Display name for the API key.",
				Required:            true,
			},
			"key":          apiKeyComputedSensitiveStringAttribute("Sensitive API key value returned by GraphOS only at creation time. Imported resources and refreshes do not rehydrate this field."),
			"created_at":   apiKeyComputedStringAttribute("Timestamp when GraphOS created the API key."),
			"last_used_at": apiKeyComputedStringAttribute("Timestamp when GraphOS last observed this API key being used, when available."),
			"id":           scaffoldResourceIDAttribute("Remote GraphOS API key identifier. Import format: `graph_id:variant:subgraph_name:key_id`."),
		},
	}
}

func (r *subgraphAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *subgraphAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subgraphAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.CreateSubgraphAPIKey(ctx, SubgraphAPIKeyInput{
		GraphID:      NormalizeString(plan.GraphID.ValueString()),
		Variant:      NormalizeString(plan.Variant.ValueString()),
		SubgraphName: NormalizeString(plan.SubgraphName.ValueString()),
		Name:         NormalizeString(plan.Name.ValueString()),
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Create subgraph API key", err, subgraphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, subgraphAPIKeyResourceModelFromAPI(key, types.StringNull()))...)
}

func (r *subgraphAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subgraphAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetSubgraphAPIKey(ctx, SubgraphAPIKeyInput{
		GraphID:      state.GraphID.ValueString(),
		Variant:      state.Variant.ValueString(),
		SubgraphName: state.SubgraphName.ValueString(),
		ID:           state.ID.ValueString(),
	})
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read subgraph API key", err, subgraphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, subgraphAPIKeyResourceModelFromAPI(key, state.Key))...)
}

func (r *subgraphAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subgraphAPIKeyResourceModel
	var state subgraphAPIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.UpdateSubgraphAPIKey(ctx, SubgraphAPIKeyInput{
		GraphID:      NormalizeString(state.GraphID.ValueString()),
		Variant:      NormalizeString(state.Variant.ValueString()),
		SubgraphName: NormalizeString(state.SubgraphName.ValueString()),
		ID:           NormalizeString(state.ID.ValueString()),
		Name:         NormalizeString(plan.Name.ValueString()),
	})
	if err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Update subgraph API key", err, subgraphAPIKeyResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, subgraphAPIKeyResourceModelFromAPI(key, state.Key))...)
}

func (r *subgraphAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subgraphAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetSubgraphAPIKey(ctx, SubgraphAPIKeyInput{
		GraphID:      state.GraphID.ValueString(),
		Variant:      state.Variant.ValueString(),
		SubgraphName: state.SubgraphName.ValueString(),
		ID:           state.ID.ValueString(),
	})
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete subgraph API key", err, subgraphAPIKeyResourceScope)
		return
	}

	err = r.client.DeleteSubgraphAPIKey(ctx, SubgraphAPIKeyInput{
		GraphID: NormalizeString(state.GraphID.ValueString()),
		ID:      NormalizeString(state.ID.ValueString()),
	})
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete subgraph API key", err, subgraphAPIKeyResourceScope)
	}
}

func (r *subgraphAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 4)
	if err != nil {
		resp.Diagnostics.AddError("Import subgraph API key", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variant"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subgraph_name"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[3])...)
}

func subgraphAPIKeyResourceModelFromAPI(key *SubgraphAPIKey, previousToken types.String) subgraphAPIKeyResourceModel {
	return subgraphAPIKeyResourceModel{
		ID:           types.StringValue(NormalizeString(key.ID)),
		GraphID:      types.StringValue(NormalizeString(key.GraphID)),
		Variant:      types.StringValue(NormalizeString(key.Variant)),
		SubgraphName: types.StringValue(NormalizeString(key.SubgraphName)),
		Name:         types.StringValue(NormalizeString(key.Name)),
		Key:          preserveAPIKeyToken(previousToken, key.Token),
		CreatedAt:    nullableStringValue(key.CreatedAt),
		LastUsedAt:   nullableStringValue(key.LastUsedAt),
	}
}
