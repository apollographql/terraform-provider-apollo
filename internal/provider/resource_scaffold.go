// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &scaffoldResource{}
var _ resource.ResourceWithImportState = &scaffoldResource{}

type scaffoldResource struct {
	client              *Client
	markdownDescription string
	schema              rschema.Schema
	typeName            string
}

func newScaffoldResource(typeName string, description string, attributes map[string]rschema.Attribute) resource.Resource {
	return &scaffoldResource{
		typeName:            typeName,
		markdownDescription: description,
		schema: rschema.Schema{
			MarkdownDescription: description,
			Attributes:          attributes,
		},
	}
}

func (r *scaffoldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.typeName
}

func (r *scaffoldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema
}

func (r *scaffoldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics, "resource")
}

func (r *scaffoldResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	addScaffoldNotImplementedError(&resp.Diagnostics, providerTypeName+"_"+r.typeName)
}

func (r *scaffoldResource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	addScaffoldNotImplementedError(&resp.Diagnostics, providerTypeName+"_"+r.typeName)
}

func (r *scaffoldResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	addScaffoldNotImplementedError(&resp.Diagnostics, providerTypeName+"_"+r.typeName)
}

func (r *scaffoldResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	addScaffoldNotImplementedError(&resp.Diagnostics, providerTypeName+"_"+r.typeName)
}

func (r *scaffoldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func NewProposalConfigResource() resource.Resource {
	return newScaffoldResource(
		"proposal_config",
		"Manages durable GraphOS proposal governance settings. The exact API shape still needs to be mapped before implementation starts.",
		map[string]rschema.Attribute{
			"graph_id": rschema.StringAttribute{
				MarkdownDescription: "Graph identifier that owns the proposal policy.",
				Required:            true,
			},
			"variant": rschema.StringAttribute{
				MarkdownDescription: "Variant targeted by the proposal configuration.",
				Required:            true,
			},
			"enabled": rschema.BoolAttribute{
				MarkdownDescription: "Whether proposal workflows are enabled for the target graph variant.",
				Optional:            true,
			},
			"require_approvals": rschema.BoolAttribute{
				MarkdownDescription: "Whether approvals should be required before proposal promotion once implemented.",
				Optional:            true,
			},
			"id": scaffoldResourceIDAttribute("Resource identifier. Intended import format: `graph_id:variant`."),
		},
	)
}

func NewSessionPolicyResource() resource.Resource {
	return newScaffoldResource(
		"session_policy",
		"Manages organization-level GraphOS session policy as a long-lived governance setting.",
		map[string]rschema.Attribute{
			"organization_id": rschema.StringAttribute{
				MarkdownDescription: "Organization identifier that owns the policy.",
				Required:            true,
			},
			"max_session_length_minutes": rschema.Int64Attribute{
				MarkdownDescription: "Maximum allowed session duration in minutes.",
				Optional:            true,
			},
			"id": scaffoldResourceIDAttribute("Resource identifier. Intended import format: `organization_id`."),
		},
	)
}

func scaffoldResourceIDAttribute(description string) rschema.StringAttribute {
	return rschema.StringAttribute{
		MarkdownDescription: description,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func configureClient(providerData any, diags *diag.Diagnostics, component string) *Client {
	if providerData == nil {
		return nil
	}

	client, ok := providerData.(*Client)
	if !ok {
		diags.AddError(
			fmt.Sprintf("Unexpected %s Configure Type", component),
			fmt.Sprintf("Expected *provider.Client, got %T. This is a provider bug.", providerData),
		)

		return nil
	}

	return client
}

func addScaffoldNotImplementedError(diags *diag.Diagnostics, subjectName string) {
	detail := "The provider surface for this object is registered from the GraphOS product specs, but the GraphOS Platform API client, CRUD wiring, import parsing, and normalization logic still need to be implemented."

	resourceType := strings.TrimPrefix(subjectName, providerTypeName+"_")
	if capability, ok := ResourceCapabilityForType(resourceType); ok {
		detail = fmt.Sprintf("%s Current capability plan: create=%s, read=%s, update=%s, delete=%s, import=%s. %s",
			detail,
			capability.Create,
			capability.Read,
			capability.Update,
			capability.Delete,
			capability.Import,
			capability.Notes,
		)
	}

	diags.AddError(
		fmt.Sprintf("Scaffolded resource not implemented: %s", subjectName),
		detail,
	)
}
