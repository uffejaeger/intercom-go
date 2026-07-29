package intercom

import (
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultRetryMaxAttempts    = 3
	defaultRetryInitialBackoff = 100 * time.Millisecond
	defaultRetryMaxBackoff     = 2 * time.Second
	defaultRetryJitter         = 0.2
	rateLimitResetHeader       = "X-RateLimit-Reset"
)

// RetryConfig controls opt-in retry behavior for transient failures and rate limits.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts, including the first request. Defaults to 3.
	MaxAttempts int
	// InitialBackoff is the first exponential backoff delay when no Retry-After hint exists. Defaults to 100ms.
	InitialBackoff time.Duration
	// MaxBackoff caps exponential backoff delays when no Retry-After hint exists. Defaults to 2s.
	MaxBackoff time.Duration
	// Jitter applies +/- this fraction to exponential backoff delays. Defaults to 0.2.
	Jitter float64
	// AllowUnsafeMethods allows retries for mutating methods such as POST, PUT, PATCH, and DELETE.
	AllowUnsafeMethods bool
	// StatusCodes overrides the HTTP status codes retried by the policy. Defaults to 408, 429, 500, 502, 503, and 504.
	StatusCodes []int
}

type retryTransport struct {
	base   http.RoundTripper
	config retryConfig
}

type retryConfig struct {
	maxAttempts       int
	initialBackoff    time.Duration
	maxBackoff        time.Duration
	jitter            float64
	allowUnsafeMethod bool
	statusCodes       map[int]bool
}

func retryHTTPClient(httpClient *http.Client, config RetryConfig) *http.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	clone := *httpClient
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = &retryTransport{
		base:   base,
		config: normalizeRetryConfig(config),
	}
	return &clone
}

func normalizeRetryConfig(config RetryConfig) retryConfig {
	maxAttempts := config.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = defaultRetryMaxAttempts
	}
	initialBackoff := config.InitialBackoff
	if initialBackoff == 0 {
		initialBackoff = defaultRetryInitialBackoff
	}
	maxBackoff := config.MaxBackoff
	if maxBackoff == 0 {
		maxBackoff = defaultRetryMaxBackoff
	}
	jitter := config.Jitter
	if jitter == 0 {
		jitter = defaultRetryJitter
	}
	if maxBackoff < initialBackoff {
		maxBackoff = initialBackoff
	}

	statusCodes := map[int]bool{}
	codes := config.StatusCodes
	if len(codes) == 0 {
		codes = []int{
			http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}
	}
	for _, code := range codes {
		statusCodes[code] = true
	}

	return retryConfig{
		maxAttempts:       maxAttempts,
		initialBackoff:    initialBackoff,
		maxBackoff:        maxBackoff,
		jitter:            jitter,
		allowUnsafeMethod: config.AllowUnsafeMethods,
		statusCodes:       statusCodes,
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	config := t.configForRequest(req)
	if config.maxAttempts <= 1 || !canRetryRequest(req, config) {
		return t.base.RoundTrip(withRetryAttempt(req, 1, config.maxAttempts))
	}

	attemptReq := req
	for attempt := 1; ; attempt++ {
		attemptReq = withRetryAttempt(attemptReq, attempt, config.maxAttempts)
		res, err := t.base.RoundTrip(attemptReq)
		if !shouldRetry(req, res, err, config) || attempt == config.maxAttempts {
			return res, err
		}

		delay := retryDelay(config, attempt, res)
		closeResponseBody(res)
		if err := sleepWithContext(req, delay); err != nil {
			return nil, err
		}

		nextReq, err := cloneRequestForRetry(req)
		if err != nil {
			return res, err
		}
		attemptReq = nextReq
	}
}

func (t *retryTransport) configForRequest(req *http.Request) retryConfig {
	if req != nil {
		if options, ok := requestOptionsFromContext(req.Context()); ok && options.Retry != nil {
			return normalizeRetryConfig(*options.Retry)
		}
	}
	return t.config
}

func canRetryRequest(req *http.Request, config retryConfig) bool {
	if req == nil {
		return false
	}
	if !config.allowUnsafeMethod && !isSafeMethod(req.Method) {
		return false
	}
	return req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
}

func shouldRetry(req *http.Request, res *http.Response, err error, config retryConfig) bool {
	if req.Context().Err() != nil {
		return false
	}
	if err != nil {
		return isRetryableNetworkError(err)
	}
	return res != nil && config.statusCodes[res.StatusCode]
}

func retryDelay(config retryConfig, attempt int, res *http.Response) time.Duration {
	if retryAfter, ok := retryAfterDelay(res); ok {
		return retryAfter
	}
	if rateLimitReset, ok := rateLimitResetDelay(res); ok {
		return rateLimitReset
	}

	exponent := min(max(attempt-1, 0), 30)

	multiplier := math.Pow(2, float64(exponent))
	delay := min(time.Duration(float64(config.initialBackoff)*multiplier), config.maxBackoff)
	if config.jitter > 0 && delay > 0 {
		factor := 1 - config.jitter + rand.Float64()*2*config.jitter
		delay = time.Duration(float64(delay) * factor)
		return min(delay, config.maxBackoff)
	}
	return delay
}

func validateRetryConfig(config RetryConfig) error {
	if config.MaxAttempts < 0 {
		return errors.New("intercom: retry max attempts cannot be negative")
	}
	if config.InitialBackoff < 0 {
		return errors.New("intercom: retry initial backoff cannot be negative")
	}
	if config.MaxBackoff < 0 {
		return errors.New("intercom: retry max backoff cannot be negative")
	}
	if config.Jitter < 0 {
		return errors.New("intercom: retry jitter cannot be negative")
	}
	if config.Jitter > 1 {
		return errors.New("intercom: retry jitter cannot be greater than 1")
	}
	return nil
}

type retryAttemptContextKey struct{}

type retryAttemptInfo struct {
	attempt     int
	maxAttempts int
}

func withRetryAttempt(req *http.Request, attempt, maxAttempts int) *http.Request {
	if req == nil {
		return nil
	}
	info := retryAttemptInfo{attempt: attempt, maxAttempts: max(maxAttempts, 1)}
	return req.WithContext(context.WithValue(req.Context(), retryAttemptContextKey{}, info))
}

func retryAttemptFromContext(ctx context.Context) retryAttemptInfo {
	if ctx == nil {
		return retryAttemptInfo{attempt: 1, maxAttempts: 1}
	}
	if info, ok := ctx.Value(retryAttemptContextKey{}).(retryAttemptInfo); ok {
		return info
	}
	return retryAttemptInfo{attempt: 1, maxAttempts: 1}
}

func cloneRequestForRetry(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body == nil || req.Body == http.NoBody {
		clone.Body = req.Body
		return clone, nil
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	clone.Body = body
	return clone, nil
}

func retryAfterDelay(res *http.Response) (time.Duration, bool) {
	if res == nil {
		return 0, false
	}
	value := strings.TrimSpace(headerValue(res.Header, "Retry-After"))
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	delay := time.Until(when)
	if delay < 0 {
		return 0, true
	}
	return delay, true
}

func rateLimitResetDelay(res *http.Response) (time.Duration, bool) {
	if res == nil || res.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}
	value := strings.TrimSpace(headerValue(res.Header, rateLimitResetHeader))
	if value == "" {
		return 0, false
	}
	resetAt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	delay := time.Until(time.Unix(resetAt, 0))
	if delay < 0 {
		return 0, true
	}
	return delay, true
}

func sleepWithContext(req *http.Request, delay time.Duration) error {
	if delay <= 0 {
		return req.Context().Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-req.Context().Done():
		return req.Context().Err()
	}
}

func closeResponseBody(res *http.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func isRetryableNetworkError(err error) bool {
	var timeoutErr interface{ Timeout() bool }
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		return true
	}
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func headerValue(header http.Header, name string) string {
	if value := header.Get(name); value != "" {
		return value
	}
	for key, values := range header {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
