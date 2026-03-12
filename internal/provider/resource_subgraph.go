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
	subgraphReadResourceScope   = "Graph read or broader org access for subgraph metadata"
	subgraphDeleteResourceScope = "Graph admin or broader org access for subgraph removal"
)

var _ resource.Resource = &subgraphResource{}
var _ resource.ResourceWithImportState = &subgraphResource{}

type subgraphResource struct {
	client *Client
}

type subgraphResourceModel struct {
	ID         types.String `tfsdk:"id"`
	GraphID    types.String `tfsdk:"graph_id"`
	Variant    types.String `tfsdk:"variant"`
	Name       types.String `tfsdk:"name"`
	RoutingURL types.String `tfsdk:"routing_url"`
	Revision   types.String `tfsdk:"revision"`
}

func NewSubgraphResource() resource.Resource {
	return &subgraphResource{}
}

func (r *subgraphResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph"
}

func (r *subgraphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Represents an existing GraphOS subgraph registration for a specific variant. New subgraphs are created in GraphOS by publishing schema, so this resource supports brownfield import, read, and delete but does not bootstrap or update registrations.",
		Attributes: map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Owning graph identifier.",
				Required:            true,
				PlanModifiers:       subgraphReplaceModifiers(),
			},
			"variant": rschema.StringAttribute{
				MarkdownDescription: "Variant that owns the subgraph registration.",
				Required:            true,
				PlanModifiers:       subgraphReplaceModifiers(),
			},
			"name": rschema.StringAttribute{
				MarkdownDescription: "Subgraph name.",
				Required:            true,
				PlanModifiers:       subgraphReplaceModifiers(),
			},
			"routing_url": rschema.StringAttribute{
				MarkdownDescription: "Routing URL returned by GraphOS for the registered subgraph.",
				Computed:            true,
			},
			"revision": rschema.StringAttribute{
				MarkdownDescription: "Current GraphOS revision identifier for the registered subgraph, when available.",
				Computed:            true,
			},
			"id": scaffoldResourceIDAttribute("Synthetic stable identifier in the form `graph_id:variant:name`. Import format: `graph_id:variant:subgraph_name`."),
		},
	}
}

func (r *subgraphResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *subgraphResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	AddAPIErrorDiagnostic(
		&resp.Diagnostics,
		"Create subgraph",
		capabilityDeferredError(
			operationPublishSubgraphSchema,
			"GraphOS creates subgraph registrations by publishing schema. `apollo_subgraph` does not model that bootstrap workflow yet, so create is deferred. Import an existing subgraph with `graph_id:variant:subgraph_name` instead.",
		),
		"",
	)
}

func (r *subgraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subgraphResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subgraph, err := r.client.GetSubgraphMetadata(ctx, state.GraphID.ValueString(), state.Variant.ValueString(), state.Name.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Read subgraph", err, subgraphReadResourceScope)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, subgraphResourceModelFromAPI(subgraph))...)
}

func (r *subgraphResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update subgraph: provider bug",
		"`apollo_subgraph` is replacement-only in this provider slice. Terraform should have planned a replace instead of calling Update.",
	)
}

func (r *subgraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subgraphResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetSubgraphMetadata(ctx, state.GraphID.ValueString(), state.Variant.ValueString(), state.Name.ValueString())
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return
		}

		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete subgraph", err, subgraphDeleteResourceScope)
		return
	}

	if err := r.client.RemoveSubgraph(ctx, state.GraphID.ValueString(), state.Variant.ValueString(), state.Name.ValueString()); err != nil {
		AddAPIErrorDiagnostic(&resp.Diagnostics, "Delete subgraph", err, subgraphDeleteResourceScope)
	}
}

func (r *subgraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := ParseImportID(req.ID, 3)
	if err != nil {
		resp.Diagnostics.AddError("Import subgraph", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variant"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[2])...)
}

func subgraphResourceModelFromAPI(subgraph *SubgraphMetadata) subgraphResourceModel {
	return subgraphResourceModel{
		ID:         types.StringValue(JoinImportID(subgraph.GraphID, subgraph.Variant, subgraph.Name)),
		GraphID:    types.StringValue(NormalizeString(subgraph.GraphID)),
		Variant:    types.StringValue(NormalizeString(subgraph.Variant)),
		Name:       types.StringValue(NormalizeString(subgraph.Name)),
		RoutingURL: nullableStringValue(NormalizeURLString(subgraph.RoutingURL)),
		Revision:   nullableStringValue(subgraph.Revision),
	}
}

func subgraphReplaceModifiers() []planmodifier.String {
	return []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
}
