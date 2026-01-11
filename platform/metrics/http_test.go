package metrics

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPTransporter implements kratoshttp.Transporter for testing
type mockHTTPTransporter struct {
	operation    string
	pathTemplate string
	request      *http.Request
	header       *mockHeader
}

func (m *mockHTTPTransporter) Kind() transport.Kind {
	return transport.KindHTTP
}

func (m *mockHTTPTransporter) Endpoint() string {
	return "localhost:8000"
}

func (m *mockHTTPTransporter) Operation() string {
	return m.operation
}

func (m *mockHTTPTransporter) RequestHeader() transport.Header {
	return m.header
}

func (m *mockHTTPTransporter) ReplyHeader() transport.Header {
	return m.header
}

func (m *mockHTTPTransporter) Request() *http.Request {
	return m.request
}

func (m *mockHTTPTransporter) PathTemplate() string {
	return m.pathTemplate
}

type mockHeader struct {
	headers map[string]string
}

func newMockHeader() *mockHeader {
	return &mockHeader{headers: make(map[string]string)}
}

func (h *mockHeader) Get(key string) string {
	return h.headers[key]
}

func (h *mockHeader) Set(key, value string) {
	h.headers[key] = value
}

func (h *mockHeader) Keys() []string {
	keys := make([]string, 0, len(h.headers))
	for k := range h.headers {
		keys = append(keys, k)
	}
	return keys
}

func (h *mockHeader) Values(key string) []string {
	if v, ok := h.headers[key]; ok {
		return []string{v}
	}
	return nil
}

func (h *mockHeader) Add(key, value string) {
	h.headers[key] = value
}

func TestHTTPMiddleware(t *testing.T) {
	reg := NewRegistry("test_service")
	mw := HTTPMiddleware(reg)

	require.NotNil(t, mw)

	// Create mock HTTP request
	httpReq, _ := http.NewRequest("GET", "/api/users/123", nil)
	tr := &mockHTTPTransporter{
		operation:    "/api.UserService/GetUser",
		pathTemplate: "/api/users/{id}",
		request:      httpReq,
		header:       newMockHeader(),
	}

	// Create context with transport
	ctx := transport.NewServerContext(context.Background(), tr)

	// Create handler that returns successfully
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// Wrap handler with middleware
	wrappedHandler := mw(handler)

	// Execute request
	resp, err := wrappedHandler(ctx, "request")
	require.NoError(t, err)
	assert.Equal(t, "response", resp)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	// Find HTTP metrics
	var foundCounter, foundHistogram bool
	for _, m := range metrics {
		if m.GetName() == "test_service_http_requests_total" {
			foundCounter = true
			assert.Equal(t, 1, len(m.GetMetric()))
			// Check labels
			labels := m.GetMetric()[0].GetLabel()
			assert.Equal(t, 3, len(labels)) // method, route, status
		}
		if m.GetName() == "test_service_http_request_duration_seconds" {
			foundHistogram = true
		}
	}

	assert.True(t, foundCounter, "Should have HTTP requests counter")
	assert.True(t, foundHistogram, "Should have HTTP request duration histogram")
}

func TestHTTPMiddleware_RoutePattern(t *testing.T) {
	tests := []struct {
		name          string
		pathTemplate  string
		operation     string
		expectedRoute string
	}{
		{
			name:          "with path template",
			pathTemplate:  "/api/users/{id}",
			operation:     "/api.UserService/GetUser",
			expectedRoute: "/api/users/{id}",
		},
		{
			name:          "without path template",
			pathTemplate:  "",
			operation:     "/api.UserService/GetUser",
			expectedRoute: "/api.UserService/GetUser",
		},
		{
			name:          "empty operation",
			pathTemplate:  "",
			operation:     "",
			expectedRoute: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new registry for each test to avoid conflicts
			reg := NewRegistry("test_service_" + tt.name)
			mw := HTTPMiddleware(reg)

			httpReq, _ := http.NewRequest("GET", "/api/users/123", nil)
			tr := &mockHTTPTransporter{
				operation:    tt.operation,
				pathTemplate: tt.pathTemplate,
				request:      httpReq,
				header:       newMockHeader(),
			}

			ctx := transport.NewServerContext(context.Background(), tr)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "ok", nil
			}

			wrappedHandler := mw(handler)
			_, err := wrappedHandler(ctx, "req")
			require.NoError(t, err)

			// Verify route label
			metrics, err := reg.Unwrap().Gather()
			require.NoError(t, err)

			for _, m := range metrics {
				if m.GetName() == "test_service_"+tt.name+"_http_requests_total" {
					labels := m.GetMetric()[0].GetLabel()
					for _, label := range labels {
						if label.GetName() == "route" {
							assert.Equal(t, tt.expectedRoute, label.GetValue())
						}
					}
				}
			}
		})
	}
}

func TestHTTPMiddleware_Methods(t *testing.T) {
	reg := NewRegistry("test_service")
	mw := HTTPMiddleware(reg)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		httpReq, _ := http.NewRequest(method, "/api/resource", nil)
		tr := &mockHTTPTransporter{
			operation:    "/api/resource",
			pathTemplate: "/api/resource",
			request:      httpReq,
			header:       newMockHeader(),
		}

		ctx := transport.NewServerContext(context.Background(), tr)

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		wrappedHandler := mw(handler)
		_, err := wrappedHandler(ctx, "req")
		require.NoError(t, err)
	}

	// Verify all methods are recorded
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_http_requests_total" {
			assert.Equal(t, len(methods), len(m.GetMetric()))
		}
	}
}

func TestHTTPMiddleware_ErrorStatus(t *testing.T) {
	reg := NewRegistry("test_service")
	mw := HTTPMiddleware(reg)

	httpReq, _ := http.NewRequest("GET", "/api/error", nil)
	tr := &mockHTTPTransporter{
		operation:    "/api/error",
		pathTemplate: "/api/error",
		request:      httpReq,
		header:       newMockHeader(),
	}

	ctx := transport.NewServerContext(context.Background(), tr)

	// Handler that returns error
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	wrappedHandler := mw(handler)
	_, err := wrappedHandler(ctx, "req")
	require.Error(t, err)

	// Check metrics recorded with error status
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_http_requests_total" {
			labels := m.GetMetric()[0].GetLabel()
			for _, label := range labels {
				if label.GetName() == "status" {
					assert.Equal(t, "500", label.GetValue())
				}
			}
		}
	}
}

func TestHTTPMiddleware_NilRegistry(t *testing.T) {
	mw := HTTPMiddleware(nil)
	require.NotNil(t, mw)

	// Should return a no-op middleware
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	wrappedHandler := mw(handler)
	resp, err := wrappedHandler(context.Background(), "req")
	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestHTTPMiddleware_NoTransport(t *testing.T) {
	reg := NewRegistry("test_service")
	mw := HTTPMiddleware(reg)

	// Context without transport
	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	wrappedHandler := mw(handler)
	resp, err := wrappedHandler(ctx, "req")
	require.NoError(t, err)
	assert.Equal(t, "ok", resp)

	// Should not record metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_http_requests_total" {
			assert.Equal(t, 0, len(m.GetMetric()))
		}
	}
}

func TestHTTPMiddleware_DurationRecording(t *testing.T) {
	reg := NewRegistry("test_service")
	mw := HTTPMiddleware(reg)

	httpReq, _ := http.NewRequest("GET", "/api/slow", nil)
	tr := &mockHTTPTransporter{
		operation:    "/api/slow",
		pathTemplate: "/api/slow",
		request:      httpReq,
		header:       newMockHeader(),
	}

	ctx := transport.NewServerContext(context.Background(), tr)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Simulate some work
		return "ok", nil
	}

	wrappedHandler := mw(handler)
	_, err := wrappedHandler(ctx, "req")
	require.NoError(t, err)

	// Check histogram was recorded
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_http_request_duration_seconds" {
			assert.Greater(t, len(m.GetMetric()), 0, "Should have recorded duration")
			// Check histogram has count
			assert.Greater(t, m.GetMetric()[0].GetHistogram().GetSampleCount(), uint64(0))
		}
	}
}
