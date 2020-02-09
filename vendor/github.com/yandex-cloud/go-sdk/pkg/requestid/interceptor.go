package requestid

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	clientTraceIDHeader   = "x-client-trace-id"
	clientRequestIDHeader = "x-client-request-id"
	serverRequestIDHeader = "x-request-id"
	serverTraceIDHeader   = "x-server-trace-id"
)

func Interceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		clientTraceID := uuid.New().String()
		clientRequestID := uuid.New().String()

		md, ok := metadata.FromOutgoingContext(ctx)
		if ok && len(md.Get(clientTraceIDHeader)) > 0 {
			clientTraceID = md.Get(clientTraceIDHeader)[0]
		}

		ctx = withMetadata(ctx, map[string]string{
			clientRequestIDHeader: clientRequestID,
			clientTraceIDHeader:   clientTraceID,
		})

		var responseHeader metadata.MD
		opts = append(opts, grpc.Header(&responseHeader))
		err := invoker(ctx, method, req, reply, conn, opts...)
		return wrapError(err, clientTraceID, clientRequestID, responseHeader)
	}
}

type RequestIDs struct {
	ClientTraceID   string
	ClientRequestID string
	ServerRequestID string
	ServerTraceID   string
}

type errorWithRequestIDs struct {
	origErr error
	ids     RequestIDs
}

func (e *errorWithRequestIDs) Error() (msg string) {
	if e.ids.ServerRequestID != "" {
		msg += fmt.Sprintf("server-request-id = %s ", e.ids.ServerRequestID)
	}
	if e.ids.ClientRequestID != "" {
		msg += fmt.Sprintf("client-request-id = %s ", e.ids.ClientRequestID)
	}
	if e.ids.ClientTraceID != "" {
		msg += fmt.Sprintf("client-trace-id = %s ", e.ids.ClientTraceID)
	}
	return msg + e.origErr.Error()
}

func (e errorWithRequestIDs) GRPCStatus() *status.Status {
	return status.Convert(e.origErr)
}

func RequestIDsFromError(err error) (*RequestIDs, bool) {
	if withID, ok := err.(*errorWithRequestIDs); ok {
		return &withID.ids, ok
	}
	return nil, false
}

func ContextWithClientTraceID(ctx context.Context, clientTraceID string) context.Context {
	return withMetadata(ctx, map[string]string{
		clientTraceIDHeader: clientTraceID,
	})
}

func wrapError(err error, clientTraceID, clientRequestID string, responseHeader metadata.MD) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*errorWithRequestIDs); ok {
		return err
	}

	serverRequestID := getServerHeader(responseHeader, serverRequestIDHeader)
	serverTraceID := getServerHeader(responseHeader, serverTraceIDHeader)

	return &errorWithRequestIDs{
		err,
		RequestIDs{
			ClientTraceID:   clientTraceID,
			ClientRequestID: clientRequestID,
			ServerRequestID: serverRequestID,
			ServerTraceID:   serverTraceID,
		},
	}
}

func getServerHeader(responseHeader metadata.MD, key string) string {
	serverHeaderIDRaw := responseHeader.Get(key)
	if len(serverHeaderIDRaw) == 0 {
		return ""
	}

	return serverHeaderIDRaw[0]
}

func withMetadata(ctx context.Context, meta map[string]string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy()
	}
	for k, v := range meta {
		md.Set(k, v)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
