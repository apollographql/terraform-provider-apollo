// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestApolloProviderMetadata(t *testing.T) {
	t.Parallel()

	p := &ApolloProvider{version: "test"}

	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != providerTypeName {
		t.Fatalf("expected provider type %q, got %q", providerTypeName, resp.TypeName)
	}

	if resp.Version != "test" {
		t.Fatalf("expected provider version %q, got %q", "test", resp.Version)
	}
}

func TestResolveProviderConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		model          ApolloProviderModel
		env            map[string]string
		expectedConfig resolvedProviderConfig
		expectedError  string
	}{
		{
			name: "explicit config wins",
			model: ApolloProviderModel{
				APIKey:   types.StringValue("service:key"),
				Endpoint: types.StringValue("https://custom.example.com/graphql"),
			},
			env: map[string]string{
				apiKeyEnvVar:   "env:key",
				endpointEnvVar: "https://env.example.com/graphql",
			},
			expectedConfig: resolvedProviderConfig{
				APIKey:   "service:key",
				Endpoint: "https://custom.example.com/graphql",
			},
		},
		{
			name: "environment fallback",
			model: ApolloProviderModel{
				APIKey:   types.StringNull(),
				Endpoint: types.StringNull(),
			},
			env: map[string]string{
				apiKeyEnvVar:   "env:key",
				endpointEnvVar: "https://env.example.com/graphql",
			},
			expectedConfig: resolvedProviderConfig{
				APIKey:   "env:key",
				Endpoint: "https://env.example.com/graphql",
			},
		},
		{
			name: "default endpoint",
			model: ApolloProviderModel{
				APIKey:   types.StringValue("service:key"),
				Endpoint: types.StringNull(),
			},
			expectedConfig: resolvedProviderConfig{
				APIKey:   "service:key",
				Endpoint: defaultEndpoint,
			},
		},
		{
			name: "unknown api key",
			model: ApolloProviderModel{
				APIKey:   types.StringUnknown(),
				Endpoint: types.StringNull(),
			},
			expectedError: "Unknown Apollo API Key",
		},
		{
			name: "unknown endpoint",
			model: ApolloProviderModel{
				APIKey:   types.StringValue("service:key"),
				Endpoint: types.StringUnknown(),
			},
			expectedError: "Unknown GraphOS Endpoint",
		},
		{
			name: "missing key",
			model: ApolloProviderModel{
				APIKey:   types.StringNull(),
				Endpoint: types.StringNull(),
			},
			expectedConfig: resolvedProviderConfig{
				Endpoint: defaultEndpoint,
			},
			expectedError: "Missing Apollo API Key",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config, diagnostics := resolveProviderConfig(testCase.model, func(key string) string {
				return testCase.env[key]
			})

			if config != testCase.expectedConfig {
				t.Fatalf("unexpected config: got %#v want %#v", config, testCase.expectedConfig)
			}

			if testCase.expectedError == "" {
				if diagnostics.HasError() {
					t.Fatalf("expected no error diagnostics, got %#v", diagnostics)
				}

				return
			}

			if !diagnosticsContainSummary(diagnostics, testCase.expectedError) {
				t.Fatalf("expected diagnostics to contain %q, got %#v", testCase.expectedError, diagnostics)
			}
		})
	}
}

func diagnosticsContainSummary(diagnostics diag.Diagnostics, summary string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Summary(), summary) {
			return true
		}
	}

	return false
}
