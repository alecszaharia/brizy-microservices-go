package metrics

import (
	"context"
	"errors"
	"platform/build"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockGRPCTransporter implements Transporter for testing gRPC
type mockGRPCTransporter struct {
	operation string
	header    *mockHeader
}

func (m *mockGRPCTransporter) Kind() transport.Kind {
	return transport.KindGRPC
}

func (m *mockGRPCTransporter) Endpoint() string {
	return "localhost:9000"
}

func (m *mockGRPCTransporter) Operation() string {
	return m.operation
}

func (m *mockGRPCTransporter) RequestHeader() transport.Header {
	return m.header
}

func (m *mockGRPCTransporter) ReplyHeader() transport.Header {
	return m.header
}

func TestGRPCMiddleware(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))
	mw := GRPCMiddleware(reg)

	require.NotNil(t, mw)

	tr := &mockGRPCTransporter{
		operation: "/api.v1.UserService/GetUser",
		header:    newMockHeader(),
	}

	ctx := transport.NewServerContext(context.Background(), tr)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	wrappedHandler := mw(handler)
	resp, err := wrappedHandler(ctx, "request")
	require.NoError(t, err)
	assert.Equal(t, "response", resp)

	// Gather metrics
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	var foundCounter, foundHistogram bool
	for _, m := range metrics {
		if m.GetName() == "test_service_grpc_requests_total" {
			foundCounter = true
			assert.Equal(t, 1, len(m.GetMetric()))
		}
		if m.GetName() == "test_service_grpc_request_duration_seconds" {
			foundHistogram = true
		}
	}

	assert.True(t, foundCounter, "Should have gRPC requests counter")
	assert.True(t, foundHistogram, "Should have gRPC request duration histogram")
}

func TestGRPCMiddleware_StatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus string
	}{
		{
			name:           "success",
			err:            nil,
			expectedStatus: "OK",
		},
		{
			name:           "not found",
			err:            status.Error(codes.NotFound, "not found"),
			expectedStatus: "NotFound",
		},
		{
			name:           "invalid argument",
			err:            status.Error(codes.InvalidArgument, "invalid"),
			expectedStatus: "InvalidArgument",
		},
		{
			name:           "internal error",
			err:            status.Error(codes.Internal, "internal error"),
			expectedStatus: "Internal",
		},
		{
			name:           "unauthenticated",
			err:            status.Error(codes.Unauthenticated, "unauthorized"),
			expectedStatus: "Unauthenticated",
		},
		{
			name:           "non-grpc error",
			err:            errors.New("regular error"),
			expectedStatus: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))
			mw := GRPCMiddleware(reg)

			tr := &mockGRPCTransporter{
				operation: "/test.Service/Method",
				header:    newMockHeader(),
			}

			ctx := transport.NewServerContext(context.Background(), tr)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "ok", tt.err
			}

			wrappedHandler := mw(handler)
			_, err := wrappedHandler(ctx, "req")
			if tt.err != nil {
				require.Error(t, err)
			}

			// Check status code in metrics
			metrics, err := reg.Unwrap().Gather()
			require.NoError(t, err)

			for _, m := range metrics {
				if m.GetName() == "test_service_grpc_requests_total" {
					labels := m.GetMetric()[0].GetLabel()
					for _, label := range labels {
						if label.GetName() == "status" {
							assert.Equal(t, tt.expectedStatus, label.GetValue())
						}
					}
				}
			}
		})
	}
}

func TestParseGRPCOperation(t *testing.T) {
	tests := []struct {
		name            string
		operation       string
		expectedService string
		expectedMethod  string
	}{
		{
			name:            "full path",
			operation:       "/api.v1.UserService/GetUser",
			expectedService: "api.v1.UserService",
			expectedMethod:  "GetUser",
		},
		{
			name:            "without leading slash",
			operation:       "api.v1.UserService/GetUser",
			expectedService: "api.v1.UserService",
			expectedMethod:  "GetUser",
		},
		{
			name:            "simple path",
			operation:       "/UserService/GetUser",
			expectedService: "UserService",
			expectedMethod:  "GetUser",
		},
		{
			name:            "no slash",
			operation:       "GetUser",
			expectedService: "unknown",
			expectedMethod:  "GetUser",
		},
		{
			name:            "empty",
			operation:       "",
			expectedService: "unknown",
			expectedMethod:  "unknown",
		},
		{
			name:            "multiple packages",
			operation:       "/com.example.api.v1.UserService/GetUser",
			expectedService: "com.example.api.v1.UserService",
			expectedMethod:  "GetUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, method := parseGRPCOperation(tt.operation)
			assert.Equal(t, tt.expectedService, service)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestGRPCMiddleware_NilRegistry(t *testing.T) {
	mw := GRPCMiddleware(nil)
	require.NotNil(t, mw)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	wrappedHandler := mw(handler)
	resp, err := wrappedHandler(context.Background(), "req")
	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestGRPCMiddleware_NoTransport(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))
	mw := GRPCMiddleware(reg)

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
		if m.GetName() == "test_service_grpc_requests_total" {
			assert.Equal(t, 0, len(m.GetMetric()))
		}
	}
}

func TestGRPCMiddleware_DurationRecording(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))
	mw := GRPCMiddleware(reg)

	tr := &mockGRPCTransporter{
		operation: "/test.Service/SlowMethod",
		header:    newMockHeader(),
	}

	ctx := transport.NewServerContext(context.Background(), tr)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	wrappedHandler := mw(handler)
	_, err := wrappedHandler(ctx, "req")
	require.NoError(t, err)

	// Check histogram was recorded
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_grpc_request_duration_seconds" {
			assert.Greater(t, len(m.GetMetric()), 0, "Should have recorded duration")
			assert.Greater(t, m.GetMetric()[0].GetHistogram().GetSampleCount(), uint64(0))
		}
	}
}

func TestGRPCMiddleware_MultipleRequests(t *testing.T) {
	reg := NewRegistry(build.NewBuildInfo("test_service", "1.0.0"))
	mw := GRPCMiddleware(reg)

	operations := []string{
		"/users.UserService/GetUser",
		"/users.UserService/CreateUser",
		"/products.ProductService/GetProduct",
	}

	for _, op := range operations {
		tr := &mockGRPCTransporter{
			operation: op,
			header:    newMockHeader(),
		}

		ctx := transport.NewServerContext(context.Background(), tr)

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		wrappedHandler := mw(handler)
		_, err := wrappedHandler(ctx, "req")
		require.NoError(t, err)
	}

	// Check all operations were recorded
	metrics, err := reg.Unwrap().Gather()
	require.NoError(t, err)

	for _, m := range metrics {
		if m.GetName() == "test_service_grpc_requests_total" {
			assert.Equal(t, len(operations), len(m.GetMetric()))
		}
	}
}
