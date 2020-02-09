package retry

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Interceptor returns retry interceptor, that can be parametrized by specific call options.
// Without any option, it uses default options, that basically retries nothing.
// Default retry quantity is 0, backoff function is nil, retry codes are DefaultRetriableCodes, AttemptHeader is false, and perCallTimeout is 0.
func Interceptor(callOpts ...grpc.CallOption) grpc.UnaryClientInterceptor {
	i := unaryInterceptor{opts: *defaultOptions.applyOptions(callOpts)}
	return i.InterceptUnary
}

type unaryInterceptor struct {
	opts options
}

func (i *unaryInterceptor) InterceptUnary(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
	opts := i.opts.applyOptions(callOpts)
	ctx = addIdempotencyToken(ctx)
	caller := grpcCaller{ctx, method, req, reply, conn, invoker, callOpts}

	// TODO(seukyaso): consider adding some configurable callbacks for notifying/logging purpose
	callContextIsDone, err := caller.Call(0, opts)

	for r := 0; opts.maxRetryCount < 0 || r < opts.maxRetryCount; r++ {
		if err == nil {
			return nil
		}

		// check for parent context errors, return if context is Cancelled or Deadline exceeded
		select {
		case <-ctx.Done():
			return contextErrorToGRPCError(ctx.Err())
		default:
		}

		// Always retry if call context is Done (cancelled or Deadline exceeded).
		// Thus, we ignore call errors in this case.
		if !callContextIsDone && !opts.isRetriable(err) {
			return err
		}

		err = opts.waitBackoff(ctx, r)

		if err != nil {
			return err
		}

		callContextIsDone, err = caller.Call(r+1, opts)
	}

	return err
}

type grpcCaller struct {
	ctx      context.Context
	method   string
	req      interface{}
	reply    interface{}
	conn     *grpc.ClientConn
	invoker  grpc.UnaryInvoker
	callOpts []grpc.CallOption
}

func (c *grpcCaller) Call(attempt int, opts *options) (callContextIsDone bool, err error) {
	callCtx := c.ctx
	var cancel context.CancelFunc

	if attempt > 0 {
		callCtx = opts.addRetryAttemptToHeader(callCtx, attempt)

		if opts.perCallTimeout > 0 {
			callCtx, cancel = context.WithTimeout(callCtx, opts.perCallTimeout)
			defer cancel()
		}
	}

	err = c.invoker(callCtx, c.method, c.req, c.reply, c.conn, c.callOpts...)

	select {
	case <-callCtx.Done():
		callContextIsDone = true
	default:
	}
	return
}

func (opts *options) applyOptions(callOpts []grpc.CallOption) *options {
	ret := *opts
	for _, opt := range callOpts {
		if do, ok := opt.(interceptorOption); ok {
			do.applyFunc(&ret)
		}
	}
	return &ret
}

func contextErrorToGRPCError(err error) error {
	switch err {
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, err.Error())
	case context.Canceled:
		return status.Error(codes.Canceled, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}

func (opts *options) waitBackoff(ctx context.Context, attempt int) error {
	if opts.backoffFunc == nil {
		return nil
	}

	waitTime := opts.backoffFunc(attempt)
	if waitTime > 0 {
		select {
		case <-ctx.Done():
			return contextErrorToGRPCError(ctx.Err())
		case <-time.After(waitTime):
		}
	}
	return nil
}

func (opts *options) isRetriable(err error) bool {
	errCode := status.Code(err)

	for _, code := range opts.retriableCodes {
		if errCode == code {
			return true
		}
	}
	return false
}

func addIdempotencyToken(ctx context.Context) context.Context {
	const idempotencyTokenMetadataKey = "idempotency-key"

	idempotencyTokenPresent := false
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		_, idempotencyTokenPresent = md[idempotencyTokenMetadataKey]
	}

	if !idempotencyTokenPresent {
		ctx = metadata.AppendToOutgoingContext(ctx, idempotencyTokenMetadataKey, uuid.New().String())
	}

	return ctx
}

func (opts *options) addRetryAttemptToHeader(ctx context.Context, attempt int) context.Context {
	const AttemptMetadataKey = "x-retry-attempt"

	if opts.addCountToHeader {
		ctx = metadata.AppendToOutgoingContext(ctx, AttemptMetadataKey, strconv.Itoa(attempt))
	}

	return ctx
}
