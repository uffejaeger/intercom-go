package intercom

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// RequestOptions configures one API request through its context.
//
// Headers and Query override values already present on the generated request.
// Retry overrides the client's retry policy; set MaxAttempts to 1 to disable
// retries for a request.
type RequestOptions struct {
	Headers http.Header
	Query   url.Values
	Retry   *RetryConfig
}

type requestOptionsContextKey struct{}

// WithRequestOptions returns a context carrying options for one API request.
//
// The returned context can be passed to any service method. Values are cloned
// so callers may safely reuse or mutate the supplied maps afterward.
func WithRequestOptions(ctx context.Context, options RequestOptions) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("intercom: request options context is nil")
	}
	if options.Retry != nil {
		if err := validateRetryConfig(*options.Retry); err != nil {
			return nil, err
		}
	}

	cloned := RequestOptions{
		Headers: options.Headers.Clone(),
		Query:   cloneURLValues(options.Query),
	}
	if options.Retry != nil {
		retry := *options.Retry
		retry.StatusCodes = append([]int(nil), options.Retry.StatusCodes...)
		cloned.Retry = &retry
	}

	return context.WithValue(ctx, requestOptionsContextKey{}, cloned), nil
}

func requestOptionsFromContext(ctx context.Context) (RequestOptions, bool) {
	if ctx == nil {
		return RequestOptions{}, false
	}
	options, ok := ctx.Value(requestOptionsContextKey{}).(RequestOptions)
	return options, ok
}

func applyRequestOptions(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}
	options, ok := requestOptionsFromContext(ctx)
	if !ok {
		return
	}

	for name, values := range options.Headers {
		req.Header.Del(name)
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	query := req.URL.Query()
	for name, values := range options.Query {
		query.Del(name)
		for _, value := range values {
			query.Add(name, value)
		}
	}
	req.URL.RawQuery = query.Encode()
}

func cloneURLValues(values url.Values) url.Values {
	if values == nil {
		return nil
	}
	clone := make(url.Values, len(values))
	for name, entries := range values {
		clone[name] = append([]string(nil), entries...)
	}
	return clone
}
