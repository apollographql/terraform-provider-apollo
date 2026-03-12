// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resolvedProviderConfig struct {
	APIKey   string
	Endpoint string
}

func validateProviderModel(model ApolloProviderModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if model.APIKey.IsUnknown() {
		diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Apollo API Key",
			"The provider cannot create a GraphOS client with an unknown `api_key` value.",
		)
	}

	if model.Endpoint.IsUnknown() {
		diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown GraphOS Endpoint",
			"The provider cannot create a GraphOS client with an unknown `endpoint` value.",
		)
	}

	return diagnostics
}

func resolveProviderConfig(model ApolloProviderModel, lookupEnv func(string) string) (resolvedProviderConfig, diag.Diagnostics) {
	diagnostics := validateProviderModel(model)
	if diagnostics.HasError() {
		return resolvedProviderConfig{}, diagnostics
	}

	config := resolvedProviderConfig{
		APIKey:   stringsFromConfigOrEnv(model.APIKey, apiKeyEnvVar, "", lookupEnv),
		Endpoint: stringsFromConfigOrEnv(model.Endpoint, endpointEnvVar, defaultEndpoint, lookupEnv),
	}

	if config.APIKey == "" {
		diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Apollo API Key",
			fmt.Sprintf("Set `api_key` in the provider configuration or export `%s` before running Terraform.", apiKeyEnvVar),
		)
	}

	return config, diagnostics
}

func stringsFromConfigOrEnv(value types.String, envVar string, fallback string, lookupEnv func(string) string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}

	if lookupEnv != nil {
		if envValue := lookupEnv(envVar); envValue != "" {
			return envValue
		}
	}

	return fallback
}
