package intercom

import (
	"errors"
	"net/http"
	"strings"
)

// Option configures a Client.
type Option func(*Client) error

// WithHTTPClient configures the HTTP client used to send requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) error {
		if httpClient == nil {
			return errors.New("intercom: http client is nil")
		}
		client.httpClient = httpClient
		return nil
	}
}

// WithBaseURL configures a custom API base URL.
func WithBaseURL(baseURL string) Option {
	return func(client *Client) error {
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if baseURL == "" {
			return errors.New("intercom: base URL is required")
		}
		client.baseURL = baseURL
		return nil
	}
}

// WithRegion configures the Intercom API region.
func WithRegion(region Region) Option {
	return func(client *Client) error {
		switch region {
		case US:
			client.baseURL = "https://api.intercom.io"
		case EU:
			client.baseURL = "https://api.eu.intercom.io"
		case AU:
			client.baseURL = "https://api.au.intercom.io"
		default:
			return errors.New("intercom: unknown region")
		}
		return nil
	}
}

// WithAPIVersion configures the Intercom-Version header.
func WithAPIVersion(version string) Option {
	return func(client *Client) error {
		version = strings.TrimSpace(version)
		if version == "" {
			return errors.New("intercom: API version is required")
		}
		client.apiVersion = version
		return nil
	}
}

// WithUserAgent configures the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(client *Client) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" {
			return errors.New("intercom: user agent is required")
		}
		client.userAgent = userAgent
		return nil
	}
}
