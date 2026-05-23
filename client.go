package intercom

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

const (
	defaultBaseURL    = "https://api.intercom.io"
	defaultAPIVersion = "2.15"
	defaultUserAgent  = "intercom-go"
)

// Region identifies an Intercom API region.
type Region string

const (
	US Region = "us"
	EU Region = "eu"
	AU Region = "au"
)

// Client is the root Intercom API client.
type Client struct {
	baseURL    string
	token      string
	apiVersion string
	userAgent  string
	httpClient *http.Client
	generated  *gen.ClientWithResponses

	Admins   *AdminsService
	Contacts *ContactsService
}

// NewClient creates an Intercom API client using bearer-token authentication.
func NewClient(token string, opts ...Option) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("intercom: token is required")
	}

	client := &Client{
		baseURL:    defaultBaseURL,
		token:      token,
		apiVersion: defaultAPIVersion,
		userAgent:  defaultUserAgent,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	generated, err := gen.NewClientWithResponses(
		client.baseURL,
		gen.WithHTTPClient(client.httpClient),
		gen.WithRequestEditorFn(client.editGeneratedRequest),
	)
	if err != nil {
		return nil, fmt.Errorf("intercom: create generated client: %w", err)
	}

	client.generated = generated
	client.Admins = &AdminsService{client: client}
	client.Contacts = &ContactsService{client: client}

	return client, nil
}

// BaseURL returns the API base URL used by the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Do sends an HTTP request with Intercom authentication and default headers.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("intercom: request is nil")
	}

	req = req.Clone(req.Context())
	c.applyDefaultHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 400 {
		return res, nil
	}

	defer res.Body.Close()
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	return nil, parseErrorResponse(res.StatusCode, body)
}

// NewRequest creates a request relative to the Intercom API base URL.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("intercom: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) editGeneratedRequest(_ context.Context, req *http.Request) error {
	c.applyDefaultHeaders(req)
	return nil
}

func (c *Client) applyDefaultHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Intercom-Version", c.apiVersion)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
}
