// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	graphAPIKeyResourceScope    = "Org Admin or Graph Admin access for graph API key management"
	subgraphAPIKeyResourceScope = "Org Admin or Graph Admin access for subgraph API key management"
)

var graphAPIKeyRoleValues = []string{
	"BILLING_MANAGER",
	"CONSUMER",
	"CONTRIBUTOR",
	"DOCUMENTER",
	"GRAPH_ADMIN",
	"LEGACY_GRAPH_KEY",
	"OBSERVER",
	"ORG_ADMIN",
	"PERSISTED_QUERY_PUBLISHER",
}

func graphAPIKeyRoleValidator() validator.String {
	return graphAPIKeyRoleStringValidator{}
}

func apiKeyReplacePlanModifiers() []planmodifier.String {
	return []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
}

func apiKeyComputedStringAttribute(description string) rschema.StringAttribute {
	return rschema.StringAttribute{
		MarkdownDescription: description,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func apiKeyComputedSensitiveStringAttribute(description string) rschema.StringAttribute {
	attribute := apiKeyComputedStringAttribute(description)
	attribute.Sensitive = true
	return attribute
}

func preserveAPIKeyToken(previous types.String, next string) types.String {
	if normalized := NormalizeString(next); normalized != "" {
		return types.StringValue(normalized)
	}

	if !previous.IsNull() && !previous.IsUnknown() {
		return types.StringValue(NormalizeString(previous.ValueString()))
	}

	return types.StringNull()
}

func validateGraphAPIKeyRoleValue(role string) error {
	normalized := NormalizeString(role)
	for _, candidate := range graphAPIKeyRoleValues {
		if normalized == candidate {
			return nil
		}
	}

	return fmt.Errorf("graph API key role must be one of %s", strings.Join(graphAPIKeyRoleValues, ", "))
}

type graphAPIKeyRoleStringValidator struct{}

func (v graphAPIKeyRoleStringValidator) Description(context.Context) string {
	return "must be one of the GraphOS UserPermission enum values"
}

func (v graphAPIKeyRoleStringValidator) MarkdownDescription(context.Context) string {
	return "must be one of the GraphOS `UserPermission` enum values"
}

func (v graphAPIKeyRoleStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if err := validateGraphAPIKeyRoleValue(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid graph API key role",
			err.Error(),
		)
	}
}
