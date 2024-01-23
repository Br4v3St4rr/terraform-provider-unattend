// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure UnattendISOProvider satisfies various provider interfaces.
var _ provider.Provider = &UnattendISOProvider{}

// UnattendISOProvider defines the provider implementation.
type UnattendISOProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// UnattendISOProviderModel describes the provider data model.
type UnattendISOProviderModel struct {
}

func (p *UnattendISOProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unattend"
	resp.Version = p.version
}

func (p *UnattendISOProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *UnattendISOProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *UnattendISOProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUnattendedISOResource,
	}
}

func (p *UnattendISOProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnattendISOProvider{
			version: version,
		}
	}
}
