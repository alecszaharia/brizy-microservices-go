package middleware

import (
	"context"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockTransporter is a mock implementation of transport.Transporter for testing
type MockTransporter struct {
	requestHeaders map[string][]string
	replyHeaders   map[string][]string
}

func newMockTransporter() *MockTransporter {
	return &MockTransporter{
		requestHeaders: make(map[string][]string),
		replyHeaders:   make(map[string][]string),
	}
}

func (m *MockTransporter) Kind() transport.Kind {
	return transport.KindHTTP
}

func (m *MockTransporter) Endpoint() string {
	return "http://localhost:8000"
}

func (m *MockTransporter) Operation() string {
	return "/api/v1/symbols"
}

func (m *MockTransporter) RequestHeader() transport.Header {
	return &mockHeader{headers: m.requestHeaders}
}

func (m *MockTransporter) ReplyHeader() transport.Header {
	return &mockHeader{headers: m.replyHeaders}
}

// mockHeader is a mock implementation of transport.Header
type mockHeader struct {
	headers map[string][]string
}

func (m *mockHeader) Get(key string) string {
	values := m.headers[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (m *mockHeader) Set(key string, value string) {
	m.headers[key] = []string{value}
}

func (m *mockHeader) Add(key string, value string) {
	m.headers[key] = append(m.headers[key], value)
}

func (m *mockHeader) Keys() []string {
	keys := make([]string, 0, len(m.headers))
	for key := range m.headers {
		keys = append(keys, key)
	}
	return keys
}

func (m *mockHeader) Values(key string) []string {
	return m.headers[key]
}

// createTestLogger creates a test logger for use in tests
func createTestLogger() log.Logger {
	logger := log.NewStdLogger(io.Discard)
	return logger
}

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		setupContext    func() context.Context
		setupRequest    interface{}
		checkContext    func(*testing.T, context.Context)
		checkTransport  func(*testing.T, *MockTransporter)
		expectHandlerOk bool
	}{
		{
			name: "generates new UUID when no X-Request-ID header",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)
				assert.IsType(t, "", requestID)

				// Verify it's a valid UUID
				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				_, err := uuid.Parse(ridStr)
				assert.NoError(t, err, "request ID should be a valid UUID")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)
				assert.NotEmpty(t, responseID[0])

				// Verify it's a valid UUID
				_, err := uuid.Parse(responseID[0])
				assert.NoError(t, err, "response request ID should be a valid UUID")
			},
			expectHandlerOk: true,
		},
		{
			name: "uses existing valid UUID X-Request-ID from request header",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				validUUID := uuid.New().String()
				mt.requestHeaders["X-Request-ID"] = []string{validUUID}
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Verify it's a valid UUID
				err := uuid.Validate(ridStr)
				assert.NoError(t, err, "should use valid UUID from header")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)

				// Should match the request header UUID
				requestUUID := mt.requestHeaders["X-Request-ID"][0]
				assert.Equal(t, requestUUID, responseID[0])
			},
			expectHandlerOk: true,
		},
		{
			name: "rejects invalid UUID and generates new one (malformed UUID)",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				mt.requestHeaders["X-Request-ID"] = []string{"not-a-valid-uuid"}
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Should be different from the invalid one
				assert.NotEqual(t, "not-a-valid-uuid", ridStr)

				// Should be a valid UUID (newly generated)
				err := uuid.Validate(ridStr)
				assert.NoError(t, err, "should generate new valid UUID for invalid input")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)

				// Response should contain new UUID, not the invalid one
				assert.NotEqual(t, "not-a-valid-uuid", responseID[0])

				err := uuid.Validate(responseID[0])
				assert.NoError(t, err)
			},
			expectHandlerOk: true,
		},
		{
			name: "rejects invalid UUID and generates new one (random string)",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				mt.requestHeaders["X-Request-ID"] = []string{"custom-request-id-12345"}
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Should not be the invalid custom string
				assert.NotEqual(t, "custom-request-id-12345", ridStr)

				// Should be a valid UUID
				err := uuid.Validate(ridStr)
				assert.NoError(t, err, "should generate new valid UUID for custom string")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)
				assert.NotEqual(t, "custom-request-id-12345", responseID[0])
			},
			expectHandlerOk: true,
		},
		{
			name: "rejects invalid UUID and generates new one (partial UUID)",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				mt.requestHeaders["X-Request-ID"] = []string{"550e8400-e29b-41d4-a716"}
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Should not be the partial UUID
				assert.NotEqual(t, "550e8400-e29b-41d4-a716", ridStr)

				// Should be a valid complete UUID
				err := uuid.Validate(ridStr)
				assert.NoError(t, err, "should generate new valid UUID for partial UUID")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)
				assert.NotEqual(t, "550e8400-e29b-41d4-a716", responseID[0])
			},
			expectHandlerOk: true,
		},
		{
			name: "generates UUID when X-Request-ID is empty string",
			setupContext: func() context.Context {
				mt := newMockTransporter()
				mt.requestHeaders["X-Request-ID"] = []string{""}
				return transport.NewServerContext(context.Background(), mt)
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Verify it's a valid UUID (generated, not empty)
				_, err := uuid.Parse(ridStr)
				assert.NoError(t, err, "should generate new UUID when header is empty")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				responseID := mt.replyHeaders["X-Request-ID"]
				assert.NotNil(t, responseID)
				assert.Len(t, responseID, 1)
				assert.NotEmpty(t, responseID[0])
			},
			expectHandlerOk: true,
		},
		{
			name: "works without transport context (generates UUID)",
			setupContext: func() context.Context {
				// Plain context without transport
				return context.Background()
			},
			setupRequest: "test-request",
			checkContext: func(t *testing.T, ctx context.Context) {
				requestID := ctx.Value(requestIDKey{})
				assert.NotNil(t, requestID)

				ridStr, ok := requestID.(string)
				assert.True(t, ok)
				assert.NotEmpty(t, ridStr)

				// Verify it's a valid UUID
				_, err := uuid.Parse(ridStr)
				assert.NoError(t, err, "should generate UUID even without transport")
			},
			checkTransport: func(t *testing.T, mt *MockTransporter) {
				// No transport to check in this case
			},
			expectHandlerOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track if handler was called
			handlerCalled := false
			var capturedContext context.Context

			// Create mock handler
			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				capturedContext = ctx
				return "response", nil
			}

			// Create middleware with test logger
			logger := createTestLogger()
			mw := RequestIDMiddleware(logger)
			handler := mw(mockHandler)

			// Setup context
			ctx := tt.setupContext()

			// Extract mock transporter before calling handler (if exists)
			var mt *MockTransporter
			if tr, ok := transport.FromServerContext(ctx); ok {
				mt = tr.(*MockTransporter)
			}

			// Execute
			resp, err := handler(ctx, tt.setupRequest)

			// Verify handler was called
			assert.True(t, handlerCalled, "handler should be called")
			assert.NoError(t, err)
			assert.Equal(t, "response", resp)

			// Check context
			if tt.checkContext != nil {
				tt.checkContext(t, capturedContext)
			}

			// Check transport headers
			if tt.checkTransport != nil && mt != nil {
				tt.checkTransport(t, mt)
			}
		})
	}
}

func TestRequestIDMiddleware_HandlerChainPropagation(t *testing.T) {
	// This test verifies that the middleware properly chains handlers
	// and propagates the request ID through multiple middleware layers

	var firstMiddlewareCalled bool
	var secondMiddlewareCalled bool
	var handlerCalled bool
	var requestIDInFirstMiddleware string
	var requestIDInSecondMiddleware string
	var requestIDInHandler string

	// Create a chain of middleware
	firstMiddleware := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			firstMiddlewareCalled = true
			if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
				requestIDInFirstMiddleware = rid
			}
			return handler(ctx, req)
		}
	}

	secondMiddleware := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			secondMiddlewareCalled = true
			if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
				requestIDInSecondMiddleware = rid
			}
			return handler(ctx, req)
		}
	}

	finalHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
			requestIDInHandler = rid
		}
		return "success", nil
	}

	// Build the middleware chain: RequestID -> First -> Second -> Handler
	logger := createTestLogger()
	handler := RequestIDMiddleware(logger)(firstMiddleware(secondMiddleware(finalHandler)))

	// Setup context with transport and valid UUID
	mt := newMockTransporter()
	validUUID := uuid.New().String()
	mt.requestHeaders["X-Request-ID"] = []string{validUUID}
	ctx := transport.NewServerContext(context.Background(), mt)

	// Execute
	resp, err := handler(ctx, "test-request")

	// Verify all handlers were called
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, firstMiddlewareCalled)
	assert.True(t, secondMiddlewareCalled)
	assert.True(t, handlerCalled)

	// Verify request ID was propagated through all layers
	assert.Equal(t, validUUID, requestIDInFirstMiddleware)
	assert.Equal(t, validUUID, requestIDInSecondMiddleware)
	assert.Equal(t, validUUID, requestIDInHandler)
}

func TestRequestIDMiddleware_ErrorPropagation(t *testing.T) {
	// Verify that errors from the handler are properly propagated
	expectedErr := assert.AnError

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}

	logger := createTestLogger()
	mw := RequestIDMiddleware(logger)
	handler := mw(mockHandler)

	mt := newMockTransporter()
	ctx := transport.NewServerContext(context.Background(), mt)

	resp, err := handler(ctx, "test-request")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Verify request ID was still set even though handler returned error
	responseID := mt.replyHeaders["X-Request-ID"]
	assert.NotNil(t, responseID)
	assert.Len(t, responseID, 1)
	assert.NotEmpty(t, responseID[0])
}

func TestRequestID_Valuer(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedResult interface{}
	}{
		{
			name: "extracts request ID from context (any format)",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), requestIDKey{}, "custom-id-format-123")
			},
			expectedResult: "custom-id-format-123",
		},
		{
			name: "returns empty string when no request ID in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedResult: "",
		},
		{
			name: "returns empty string when type assertion fails",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), requestIDKey{}, 12345) // wrong type
			},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			valuer := RequestID()

			result := valuer(ctx)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRequestIDMiddleware_ConcurrentRequests(t *testing.T) {
	// Verify that concurrent requests get different request IDs
	const numRequests = 100

	requestIDs := make(chan string, numRequests)
	done := make(chan bool, numRequests)

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
			requestIDs <- rid
		}
		done <- true
		return "success", nil
	}

	logger := createTestLogger()
	mw := RequestIDMiddleware(logger)
	handler := mw(mockHandler)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			mt := newMockTransporter()
			ctx := transport.NewServerContext(context.Background(), mt)
			_, _ = handler(ctx, "test-request")
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(requestIDs)

	// Collect all request IDs
	ids := make(map[string]bool)
	for rid := range requestIDs {
		ids[rid] = true
	}

	// Verify all request IDs are unique
	assert.Equal(t, numRequests, len(ids), "all request IDs should be unique")

	// Verify all are valid UUIDs
	for rid := range ids {
		_, err := uuid.Parse(rid)
		assert.NoError(t, err, "request ID %s should be a valid UUID", rid)
	}
}

func TestRequestIDMiddleware_ContextPropagation(t *testing.T) {
	// Verify that other context values are preserved
	type contextKey string
	const userIDKey contextKey = "user-id"

	var capturedContext context.Context

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedContext = ctx
		return "success", nil
	}

	logger := createTestLogger()
	mw := RequestIDMiddleware(logger)
	handler := mw(mockHandler)

	// Setup context with transport and additional values
	mt := newMockTransporter()
	ctx := transport.NewServerContext(context.Background(), mt)
	ctx = context.WithValue(ctx, userIDKey, "user-123")

	_, err := handler(ctx, "test-request")
	assert.NoError(t, err)

	// Verify request ID was added
	requestID := capturedContext.Value(requestIDKey{})
	assert.NotNil(t, requestID)

	// Verify other context values are preserved
	userID := capturedContext.Value(userIDKey)
	assert.Equal(t, "user-123", userID)

	// Verify transport is still accessible
	tr, ok := transport.FromServerContext(capturedContext)
	assert.True(t, ok)
	assert.NotNil(t, tr)
}

func BenchmarkRequestIDMiddleware_WithHeader(b *testing.B) {
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	logger := createTestLogger()
	mw := RequestIDMiddleware(logger)
	handler := mw(mockHandler)

	mt := newMockTransporter()
	validUUID := uuid.New().String()
	mt.requestHeaders["X-Request-ID"] = []string{validUUID}
	ctx := transport.NewServerContext(context.Background(), mt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler(ctx, "test-request")
	}
}

func BenchmarkRequestIDMiddleware_GenerateUUID(b *testing.B) {
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	logger := createTestLogger()
	mw := RequestIDMiddleware(logger)
	handler := mw(mockHandler)

	mt := newMockTransporter()
	ctx := transport.NewServerContext(context.Background(), mt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler(ctx, "test-request")
	}
}

func BenchmarkRequestID_Valuer(b *testing.B) {
	ctx := context.WithValue(context.Background(), requestIDKey{}, "550e8400-e29b-41d4-a716-446655440000")
	valuer := RequestID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = valuer(ctx)
	}
}
