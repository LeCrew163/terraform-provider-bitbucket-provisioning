package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
)

// Client wraps the generated Bitbucket API client with additional functionality
type Client struct {
	// Generated API client
	client *bitbucket.APIClient

	// Configuration
	baseURL            string
	token              string
	username           string
	password           string
	insecureSkipVerify bool
	timeout            time.Duration

	// HTTP client
	httpClient *http.Client
}

// Config holds the configuration for the Bitbucket client
type Config struct {
	BaseURL            string
	Token              string
	Username           string
	Password           string
	InsecureSkipVerify bool
	Timeout            time.Duration
}

// NewClient creates a new Bitbucket Data Center client
func NewClient(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Validate authentication
	hasToken := config.Token != ""
	hasBasicAuth := config.Username != "" && config.Password != ""

	if !hasToken && !hasBasicAuth {
		return nil, fmt.Errorf("either token or username/password must be provided")
	}

	if hasToken && hasBasicAuth {
		return nil, fmt.Errorf("only one authentication method should be provided (token or username/password)")
	}

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create HTTP client with custom transport
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	// Wrap transport with retry logic, then optionally with bearer auth.
	// The bearer wrapper sits outermost so the token is never stored in any
	// loggable map — it lives only in an unexported struct field.
	var transport http.RoundTripper = NewRetryableTransport(baseTransport)
	if config.Token != "" {
		transport = &bearerTransport{token: config.Token, base: transport}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// Create generated API client configuration.
	// The generated client appends paths like /api/latest/projects, so the server
	// URL must include the Bitbucket REST context root (/rest).
	// Full request URL = BaseURL + "/rest" + "/api/latest/projects"
	//                  = http://host:7990/rest/api/latest/projects  ✓
	apiConfig := bitbucket.NewConfiguration()
	apiConfig.Servers = bitbucket.ServerConfigurations{
		{
			URL: config.BaseURL + "/rest",
		},
	}
	apiConfig.HTTPClient = httpClient

	// Create the client
	client := &Client{
		client:             bitbucket.NewAPIClient(apiConfig),
		baseURL:            config.BaseURL,
		token:              config.Token,
		username:           config.Username,
		password:           config.Password,
		insecureSkipVerify: config.InsecureSkipVerify,
		timeout:            config.Timeout,
		httpClient:         httpClient,
	}

	return client, nil
}

// GetAPIClient returns the underlying generated API client
func (c *Client) GetAPIClient() *bitbucket.APIClient {
	return c.client
}

// NewAuthContext creates a new context with authentication headers.
// Token auth is handled by the bearerTransport at the HTTP layer, so no
// context value is needed for that path.
func (c *Client) NewAuthContext(ctx context.Context) context.Context {
	if c.token != "" {
		return ctx
	}
	return context.WithValue(ctx, bitbucket.ContextBasicAuth, bitbucket.BasicAuth{
		UserName: c.username,
		Password: c.password,
	})
}

// bearerTransport injects Authorization: Bearer on every outgoing request.
// The token is stored in an unexported field so it never appears in any
// loggable configuration map.
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(clone)
}

// TestConnection performs a connectivity check to the Bitbucket instance
// func (c *Client) TestConnection(ctx context.Context) error {
// 	authCtx := c.NewAuthContext(ctx)
// 
// 	tflog.Debug(ctx, "Testing connection to Bitbucket Data Center", map[string]interface{}{
// 		"base_url": c.baseURL,
// 	})
// 
// 	// Try to get application properties as a connectivity test
// 	_, resp, err := c.client.ApplicationPropertiesAPI.GetApplicationProperties(authCtx).Execute()
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to Bitbucket: %w", err)
// 	}
// 
// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("unexpected status code from Bitbucket: %d", resp.StatusCode)
// 	}
// 
// 	tflog.Debug(ctx, "Successfully connected to Bitbucket Data Center")
// 	return nil
// }

// GetBaseURL returns the configured base URL
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetHTTPClient returns the underlying HTTP client (for raw requests)
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// GetUsername returns the configured username
func (c *Client) GetUsername() string {
	return c.username
}

// GetPassword returns the configured password
func (c *Client) GetPassword() string {
	return c.password
}

// GetToken returns the configured token
func (c *Client) GetToken() string {
	return c.token
}
