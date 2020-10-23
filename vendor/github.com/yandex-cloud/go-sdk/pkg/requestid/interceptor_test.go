package requestid

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	clientTraceID   = "client-trace-id"
	clientRequestID = "client-request-id"
	serverRequestID = "server-request-id"
	serverTraceID   = "server-trace-id"
)

type StatusError interface {
	GRPCStatus() *status.Status
}

func clientRequestIDFromError(err error) string {
	if info, ok := RequestIDsFromError(err); ok {
		return info.ClientRequestID
	}
	return ""
}

func clientTraceIDFromError(err error) string {
	if info, ok := RequestIDsFromError(err); ok {
		return info.ClientTraceID
	}
	return ""
}

func responseHeader(serverRequestID, serverTraceID string) metadata.MD {
	return metadata.New(map[string]string{
		serverRequestIDHeader: serverRequestID,
		serverTraceIDHeader:   serverTraceID,
	})
}

func TestWrappedRequestIDs(t *testing.T) {
	t.Run("unwrap normal error", func(t *testing.T) {
		expected := fmt.Errorf("some error")
		errorInfo, ok := RequestIDsFromError(expected)
		assert.False(t, ok)
		assert.Nil(t, errorInfo)
	})
	t.Run("unwrap nil error", func(t *testing.T) {
		errorInfo, ok := RequestIDsFromError(nil)
		assert.False(t, ok)
		assert.Nil(t, errorInfo)
	})
	t.Run("wrap nil error", func(t *testing.T) {
		actual := wrapError(nil, clientTraceID, clientRequestID, nil)
		assert.Nil(t, actual)
	})
	t.Run("wrap err with client request id and nil header", func(t *testing.T) {
		err := fmt.Errorf("some error")
		actual := wrapError(err, clientTraceID, clientRequestID, nil)
		assert.Equal(t, &errorWithRequestIDs{err, RequestIDs{clientTraceID, clientRequestID, "", ""}}, actual)

		errorInfo, ok := RequestIDsFromError(actual)
		assert.True(t, ok)
		assert.Equal(t, clientRequestID, clientRequestIDFromError(actual))
		assert.Equal(t, clientTraceID, clientTraceIDFromError(actual))
		assert.Equal(t, "", errorInfo.ServerRequestID)
		assert.Equal(t, "", errorInfo.ServerTraceID)

	})
	t.Run("wrap err with client and server request id", func(t *testing.T) {
		err := fmt.Errorf("some error")
		actual := wrapError(err, clientTraceID, clientRequestID, responseHeader(serverRequestID, serverTraceID))
		assert.Equal(t, &errorWithRequestIDs{err, RequestIDs{clientTraceID, clientRequestID, serverRequestID, serverTraceID}}, actual)

		errorInfo, ok := RequestIDsFromError(actual)
		assert.True(t, ok)
		assert.Equal(t, clientRequestID, errorInfo.ClientRequestID)
		assert.Equal(t, serverRequestID, errorInfo.ServerRequestID)
		assert.Equal(t, serverTraceID, errorInfo.ServerTraceID)
	})
	t.Run("wrap err with empty header", func(t *testing.T) {
		err := fmt.Errorf("some error")
		actual := wrapError(err, clientTraceID, clientRequestID, metadata.New(map[string]string{}))
		assert.Equal(t, &errorWithRequestIDs{err, RequestIDs{clientTraceID, clientRequestID, "", ""}}, actual)

		errorInfo, ok := RequestIDsFromError(actual)
		assert.True(t, ok)
		assert.Equal(t, clientRequestID, errorInfo.ClientRequestID)
		assert.Equal(t, "", errorInfo.ServerRequestID)
		assert.Equal(t, "", errorInfo.ServerTraceID)
	})
	t.Run("wrap wrapped", func(t *testing.T) {
		err := fmt.Errorf("some error")
		wrap1 := wrapError(err, "trace1", "id1", nil)
		wrap2 := wrapError(wrap1, "trace1", "id2", nil)
		// Should keep first clientRequestID set
		assert.Equal(t, "id1", clientRequestIDFromError(wrap1))
		assert.Equal(t, "trace1", clientTraceIDFromError(wrap1))
		assert.Equal(t, "id1", clientRequestIDFromError(wrap2))
		assert.Equal(t, "trace1", clientTraceIDFromError(wrap2))
	})
}

func TestAddClientRequestID(t *testing.T) {
	t.Run("no outgoing context", func(t *testing.T) {
		ctx := withMetadata(context.Background(), map[string]string{
			clientRequestIDHeader: clientRequestID,
			clientTraceIDHeader:   clientTraceID,
		})
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   clientTraceID,
			clientRequestIDHeader: clientRequestID,
		}), md)
	})
	t.Run("with outgoing context", func(t *testing.T) {
		ctx := context.Background()
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("it-very-long-header", "foobar"))
		ctx = withMetadata(ctx, map[string]string{
			clientRequestIDHeader: clientRequestID,
			clientTraceIDHeader:   clientTraceID,
		})
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   clientTraceID,
			clientRequestIDHeader: clientRequestID,
			"it-very-long-header": "foobar",
		}), md)
	})
	t.Run("with old request-id", func(t *testing.T) {
		ctx := context.Background()
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(clientRequestIDHeader, "old"))
		ctx = withMetadata(ctx, map[string]string{
			clientRequestIDHeader: clientRequestID,
			clientTraceIDHeader:   clientTraceID,
		})
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   clientTraceID,
			clientRequestIDHeader: clientRequestID,
		}), md)
	})
	t.Run("several ids", func(t *testing.T) {
		ctx := context.Background()
		ctx1 := withMetadata(ctx, map[string]string{
			clientRequestIDHeader: "id1",
			clientTraceIDHeader:   "trace1",
		})
		md, ok := metadata.FromOutgoingContext(ctx1)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   "trace1",
			clientRequestIDHeader: "id1",
		}), md)

		ctx2 := withMetadata(ctx1, map[string]string{
			clientRequestIDHeader: "id2",
			clientTraceIDHeader:   "trace1",
		})
		md, ok = metadata.FromOutgoingContext(ctx2)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   "trace1",
			clientRequestIDHeader: "id2",
		}), md)
		// Original context not damaged
		md, ok = metadata.FromOutgoingContext(ctx1)
		require.True(t, ok)
		assert.Equal(t, metadata.New(map[string]string{
			clientTraceIDHeader:   "trace1",
			clientRequestIDHeader: "id1",
		}), md)
	})
}

func TestWrappedErrorImplGRPCStatus(t *testing.T) {
	t.Run("wrapped error impl StatusError interface", func(t *testing.T) {
		err := fmt.Errorf("some error")
		actual := wrapError(err, clientTraceID, clientRequestID, nil)
		assert.Equal(t, &errorWithRequestIDs{err, RequestIDs{clientTraceID, clientRequestID, "", ""}}, actual)
		assert.Implements(t, (*StatusError)(nil), actual)
	})
	t.Run("get status by status.FromError method", func(t *testing.T) {
		err := fmt.Errorf("some error")
		actual := wrapError(err, clientTraceID, clientRequestID, nil)
		st, ok := status.FromError(actual)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, st.Code())
	})
	t.Run("wrap status error", func(t *testing.T) {
		sErr := status.Error(codes.Aborted, "request aborted")
		actual := wrapError(sErr, clientTraceID, clientRequestID, nil)
		st, ok := status.FromError(actual)
		assert.True(t, ok)
		assert.Equal(t, "request aborted", st.Message())
		assert.Equal(t, codes.Aborted, st.Code())
	})
}
