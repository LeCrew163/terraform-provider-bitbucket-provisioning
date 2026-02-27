package provider

import (
	"context"
	"os"
	"strconv"
	"time"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &bitbucketdcProvider{}
)

// bitbucketdcProvider is the provider implementation.
type bitbucketdcProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// bitbucketdcProviderModel maps provider schema data to a Go type.
type bitbucketdcProviderModel struct {
	BaseURL            types.String `tfsdk:"base_url"`
	Token              types.String `tfsdk:"token"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	Timeout            types.Int64  `tfsdk:"timeout"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &bitbucketdcProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *bitbucketdcProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bitbucketdc"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *bitbucketdcProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Bitbucket Data Center resources.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The base URL of your Bitbucket Data Center instance (e.g., https://bitbucket.example.com). " +
					"Can also be set via the BITBUCKET_BASE_URL environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: "Personal Access Token for authentication. " +
					"Can also be set via the BITBUCKET_TOKEN environment variable. " +
					"Either token or username/password must be provided.",
				Optional:  true,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				Description: "Username for HTTP Basic Authentication. " +
					"Can also be set via the BITBUCKET_USERNAME environment variable. " +
					"Must be used together with password.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "Password for HTTP Basic Authentication. " +
					"Can also be set via the BITBUCKET_PASSWORD environment variable. " +
					"Must be used together with username.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. " +
					"Can also be set via the BITBUCKET_INSECURE_SKIP_VERIFY environment variable. " +
					"Not recommended for production use.",
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Request timeout in seconds. Defaults to 30 seconds. " +
					"Can also be set via the BITBUCKET_TIMEOUT environment variable.",
				Optional: true,
			},
		},
	}
}

// Configure prepares a Bitbucket API client for data sources and resources.
func (p *bitbucketdcProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Bitbucket Data Center provider", map[string]interface{}{
		"version": p.version,
	})

	// Retrieve provider data from configuration
	var config bitbucketdcProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If values are not specified in configuration, try environment variables
	baseURL := getConfigValue(config.BaseURL, "BITBUCKET_BASE_URL")
	token := getConfigValue(config.Token, "BITBUCKET_TOKEN")
	username := getConfigValue(config.Username, "BITBUCKET_USERNAME")
	password := getConfigValue(config.Password, "BITBUCKET_PASSWORD")

	// Parse insecure_skip_verify from environment if not set
	insecureSkipVerify := config.InsecureSkipVerify.ValueBool()
	if config.InsecureSkipVerify.IsNull() {
		if envVal := os.Getenv("BITBUCKET_INSECURE_SKIP_VERIFY"); envVal != "" {
			if val, err := strconv.ParseBool(envVal); err == nil {
				insecureSkipVerify = val
			}
		}
	}

	// Parse timeout from environment if not set
	timeout := int64(30)
	if !config.Timeout.IsNull() {
		timeout = config.Timeout.ValueInt64()
	} else if envVal := os.Getenv("BITBUCKET_TIMEOUT"); envVal != "" {
		if val, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			timeout = val
		}
	}

	// Validate required configuration
	if baseURL == "" {
		resp.Diagnostics.AddError(
			"Missing Bitbucket Base URL",
			"The provider requires a base_url to be configured. "+
				"Set the base_url attribute in the provider configuration or use the BITBUCKET_BASE_URL environment variable.",
		)
		return
	}

	// Validate authentication
	hasToken := token != ""
	hasBasicAuth := username != "" && password != ""

	if !hasToken && !hasBasicAuth {
		resp.Diagnostics.AddError(
			"Missing Authentication Credentials",
			"The provider requires authentication credentials. "+
				"Either provide a token (BITBUCKET_TOKEN) or username/password (BITBUCKET_USERNAME and BITBUCKET_PASSWORD).",
		)
		return
	}

	if hasToken && hasBasicAuth {
		resp.Diagnostics.AddError(
			"Multiple Authentication Methods",
			"Only one authentication method should be provided. "+
				"Use either token or username/password, not both.",
		)
		return
	}

	if username != "" && password == "" {
		resp.Diagnostics.AddError(
			"Missing Password",
			"When using username authentication, password must also be provided.",
		)
		return
	}

	if password != "" && username == "" {
		resp.Diagnostics.AddError(
			"Missing Username",
			"When using password authentication, username must also be provided.",
		)
		return
	}

	// Create client configuration
	clientConfig := client.Config{
		BaseURL:            baseURL,
		Token:              token,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecureSkipVerify,
		Timeout:            time.Duration(timeout) * time.Second,
	}

	// Create client
	bitbucketClient, err := client.NewClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Bitbucket Client",
			"An error occurred when creating the Bitbucket API client: "+err.Error(),
		)
		return
	}

	// TODO: Add connection test and version check once we identify the right API endpoint
	// Test connection and check version
	// if err := bitbucketClient.TestConnection(ctx); err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Connect to Bitbucket",
	// 		"Failed to connect to Bitbucket Data Center instance: "+err.Error()+
	// 			"\n\nPlease verify:\n"+
	// 			"- The base_url is correct and accessible\n"+
	// 			"- The authentication credentials are valid\n"+
	// 			"- The Bitbucket instance is running",
	// 	)
	// 	return
	// }

	// Check version compatibility
	// version, err := bitbucketClient.CheckVersion(ctx)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Incompatible Bitbucket Version",
	// 		err.Error(),
	// 	)
	// 	return
	// }

	tflog.Info(ctx, "Successfully configured Bitbucket Data Center provider", map[string]interface{}{
		"base_url":    baseURL,
		"auth_method": getAuthMethod(hasToken),
	})

	// Make the Bitbucket client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = bitbucketClient
	resp.ResourceData = bitbucketClient
}

// DataSources defines the data sources implemented in the provider.
func (p *bitbucketdcProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be added here
	}
}

// Resources defines the resources implemented in the provider.
func (p *bitbucketdcProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewRepositoryResource,
		NewProjectPermissionsResource,
		NewBranchPermissionsResource,
		NewProjectAccessKeyResource,
		NewRepositoryPermissionsResource,
		NewRepositoryAccessKeyResource,
	}
}

// getConfigValue returns the value from the config or falls back to environment variable
func getConfigValue(configValue types.String, envVar string) string {
	if !configValue.IsNull() && configValue.ValueString() != "" {
		return configValue.ValueString()
	}
	return os.Getenv(envVar)
}

// getAuthMethod returns a string describing the authentication method
func getAuthMethod(hasToken bool) string {
	if hasToken {
		return "token"
	}
	return "basic"
}
