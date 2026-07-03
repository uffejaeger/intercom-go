package intercomtest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"

	intercom "github.com/uffejaeger/intercom-go"
)

// TestingT is the subset of testing.TB used by Server.
type TestingT interface {
	Cleanup(func())
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Helper()
}

// Handler describes one expected Intercom request and the response to return.
type Handler struct {
	method   string
	path     string
	response Response
	handled  bool
}

// Response describes an HTTP response returned by a scripted route.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Request describes one request captured by Server.
type Request struct {
	Method   string
	Path     string
	RawQuery string
	Header   http.Header
	Body     []byte
}

// Server is an offline Intercom API test server.
type Server struct {
	// URL is the local base URL clients should use.
	URL string

	t      TestingT
	server *httptest.Server

	mu       sync.Mutex
	handlers []Handler
	requests []Request
}

// NewServer starts an offline Intercom API test server with scripted routes.
func NewServer(t TestingT, handlers ...Handler) *Server {
	t.Helper()

	srv := &Server{
		t:        t,
		handlers: append([]Handler(nil), handlers...),
	}

	httpServer := httptest.NewServer(http.HandlerFunc(srv.serveHTTP))
	srv.server = httpServer
	srv.URL = httpServer.URL

	t.Cleanup(func() {
		srv.server.Close()
		srv.checkExpectations()
	})

	return srv
}

// Client creates an intercom-go client configured to call this server.
func (s *Server) Client(token string, opts ...intercom.Option) (*intercom.Client, error) {
	options := []intercom.Option{
		intercom.WithBaseURL(s.URL),
		intercom.WithHTTPClient(s.server.Client()),
	}
	options = append(options, opts...)
	return intercom.NewClient(token, options...)
}

// Route scripts one expected method/path pair and response.
func Route(method, path string, response Response) Handler {
	return Handler{
		method:   method,
		path:     path,
		response: response,
	}
}

// JSON creates a JSON response. String and []byte bodies are written as-is;
// other values are JSON encoded.
func JSON(statusCode int, body any) Response {
	data, err := encodeJSONBody(body)
	if err != nil {
		data = fmt.Appendf(nil, `{"error":"intercomtest: encode JSON response: %s"}`, err.Error())
		statusCode = http.StatusInternalServerError
	}

	return Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: data,
	}
}

// Error creates an Intercom-style API error response.
func Error(statusCode int, requestID, code, message string) Response {
	return JSON(statusCode, map[string]any{
		"type":       "error.list",
		"request_id": requestID,
		"errors": []map[string]string{{
			"code":    code,
			"message": message,
		}},
	})
}

// Bytes creates a binary response with the supplied content type.
func Bytes(statusCode int, contentType string, body []byte) Response {
	return Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: append([]byte(nil), body...),
	}
}

// NoContent creates a response with no body.
func NoContent(statusCode int) Response {
	return Response{StatusCode: statusCode}
}

// Request returns a captured request by index.
func (s *Server) Request(t TestingT, index int) Request {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.requests) {
		t.Fatalf("request index %d out of range; captured %d request(s)", index, len(s.requests))
	}
	return cloneRequest(s.requests[index])
}

// Requests returns all captured requests.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests := make([]Request, 0, len(s.requests))
	for _, req := range s.requests {
		requests = append(requests, cloneRequest(req))
	}
	return requests
}

// RequestCount returns the number of captured requests.
func (s *Server) RequestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requests)
}

// JSONBody decodes a captured request body into a map.
func (r Request) JSONBody(t TestingT) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("decode request JSON body %q: %v", string(r.Body), err)
	}
	return body
}

func (s *Server) serveHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		s.t.Errorf("read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	captured := Request{
		Method:   req.Method,
		Path:     req.URL.Path,
		RawQuery: req.URL.RawQuery,
		Header:   req.Header.Clone(),
		Body:     body,
	}

	s.mu.Lock()
	s.requests = append(s.requests, captured)
	handlerIndex := s.nextHandlerIndex(req.Method, req.URL.Path)
	if handlerIndex == -1 {
		s.mu.Unlock()
		s.t.Errorf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := s.handlers[handlerIndex].response
	s.handlers[handlerIndex].handled = true
	s.mu.Unlock()

	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if response.StatusCode == 0 {
		response.StatusCode = http.StatusOK
	}
	w.WriteHeader(response.StatusCode)
	if _, err := w.Write(response.Body); err != nil {
		s.t.Errorf("write response body: %v", err)
	}
}

func (s *Server) nextHandlerIndex(method, path string) int {
	for index := range s.handlers {
		handler := &s.handlers[index]
		if handler.handled {
			continue
		}
		if handler.method == method && handler.path == path {
			return index
		}
	}
	return -1
}

func (s *Server) checkExpectations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, handler := range s.handlers {
		if !handler.handled {
			s.t.Errorf("expected request was not made: %s %s", handler.method, handler.path)
		}
	}
}

func encodeJSONBody(body any) ([]byte, error) {
	switch body := body.(type) {
	case nil:
		return nil, nil
	case string:
		return []byte(body), nil
	case []byte:
		return append([]byte(nil), body...), nil
	default:
		return json.Marshal(body)
	}
}

func cloneRequest(req Request) Request {
	return Request{
		Method:   req.Method,
		Path:     req.Path,
		RawQuery: req.RawQuery,
		Header:   req.Header.Clone(),
		Body:     append([]byte(nil), req.Body...),
	}
}
