// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	apiKeyEnvVar     = "APOLLO_KEY"
	defaultEndpoint  = "https://graphql.api.apollographql.com/api/graphql"
	endpointEnvVar   = "APOLLO_GRAPHOS_ENDPOINT"
	providerTypeName = "apollo"
)

var _ provider.Provider = &ApolloProvider{}

type ApolloProvider struct {
	version string
}

type ApolloProviderModel struct {
	APIKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *ApolloProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

func (p *ApolloProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = pschema.Schema{
		MarkdownDescription: "Manage Apollo GraphOS control-plane configuration through Terraform.",
		Attributes: map[string]pschema.Attribute{
			"api_key": pschema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("GraphOS API key. Can also be set with the `%s` environment variable.", apiKeyEnvVar),
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": pschema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("GraphOS Platform API endpoint. Can also be set with the `%s` environment variable.", endpointEnvVar),
				Optional:            true,
			},
		},
	}
}

func (p *ApolloProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ApolloProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, diagnostics := resolveProviderConfig(data, os.Getenv)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := NewClient(config.APIKey, config.Endpoint, p.userAgent())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApolloProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGraphVariantResource,
		NewSubgraphResource,
		NewPersistedQueryListResource,
		NewPersistedQueryListLinkResource,
		NewGraphAPIKeyResource,
		NewSubgraphAPIKeyResource,
		NewProposalConfigResource,
		NewSessionPolicyResource,
	}
}

func (p *ApolloProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGraphDataSource,
		NewGraphVariantDataSource,
		NewSubgraphDataSource,
		NewPersistedQueryListDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApolloProvider{
			version: version,
		}
	}
}

func (p *ApolloProvider) userAgent() string {
	return fmt.Sprintf("terraform-provider-apollo/%s", p.version)
}
