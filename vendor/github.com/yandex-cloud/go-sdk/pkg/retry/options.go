package retry

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// WithDefaultInterceptor returns interceptor that DOESN'T retry anything.
// Its possible to change its behaviour with call options.
func WithDefaultInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(Interceptor())
}

// WithMax option sets quantity of retry attempts.
// It handles negative maxRetryCount as INFINITE retries.
func WithMax(maxRetryCount int) grpc.CallOption {
	return newOption(setRetryQuantity(maxRetryCount))
}

// WithCodes overrides the whole retriable codes list.
func WithCodes(codes ...codes.Code) grpc.CallOption {
	return newOption(setRetriableCodes(codes...))
}

// WithAttemptHeader adds retry attempt number to context outgoing metadata, with key "x-retry-attempt".
func WithAttemptHeader(enable bool) grpc.CallOption {
	return newOption(setAttemptHeader(enable))
}

// WithPerCallTimeout adds timeout for retry calls.
func WithPerCallTimeout(to time.Duration) grpc.CallOption {
	return newOption(setPerCallTimeout(to))
}

// WithBackoff sets up interceptor with custom defined backoff function
func WithBackoff(f BackoffFunc) grpc.CallOption {
	return newOption(setBackoff(f))
}

// DefaultBackoff uses exponential backoff with jitter, with base = 50ms, and maximum timeout = 1 minute.
func DefaultBackoff() BackoffFunc {
	return DefaultExponentialJitterBackoff()
}

// WithDefaultExponentialJitterBackoff same as WithDefaultBackoff
func DefaultExponentialJitterBackoff() BackoffFunc {
	return BackoffExponentialWithJitter(defaultExponentialBackoffBase, defaultExponentialBackoffCap)
}

// DefaultLinearJitterBackoff uses linear backoff with base = 50ms, and jitter = +-10%
func DefaultLinearJitterBackoff() BackoffFunc {
	return BackoffLinearWithJitter(defaultLinearBackoffTimeout, defaultLinearBackoffJitter)
}

type options struct {
	maxRetryCount    int
	retriableCodes   []codes.Code
	addCountToHeader bool
	perCallTimeout   time.Duration
	backoffFunc      BackoffFunc
}

var (
	// TODO(seukyaso): Consider adding some non-zero default retry options
	DefaultRetriableCodes = []codes.Code{codes.ResourceExhausted, codes.Unavailable}

	defaultOptions = &options{
		maxRetryCount:    0,
		retriableCodes:   DefaultRetriableCodes,
		addCountToHeader: false,
		perCallTimeout:   0,
		backoffFunc:      nil,
	}
)

const (
	defaultLinearBackoffTimeout   = 50 * time.Millisecond
	defaultLinearBackoffJitter    = 0.1
	defaultExponentialBackoffBase = 50 * time.Millisecond
	defaultExponentialBackoffCap  = 1 * time.Minute
)

type applyOptionFunc func(opt *options)

type interceptorOption struct {
	grpc.EmptyCallOption
	applyFunc applyOptionFunc
}

func newOption(f applyOptionFunc) grpc.CallOption {
	return interceptorOption{applyFunc: f}
}

func setRetryQuantity(r int) applyOptionFunc {
	return func(opt *options) {
		opt.maxRetryCount = r
	}
}

func setRetriableCodes(codes ...codes.Code) applyOptionFunc {
	return func(opt *options) {
		opt.retriableCodes = codes
	}
}

func setAttemptHeader(enable bool) applyOptionFunc {
	return func(opt *options) {
		opt.addCountToHeader = enable
	}
}

func setPerCallTimeout(to time.Duration) applyOptionFunc {
	return func(opt *options) {
		opt.perCallTimeout = to
	}
}

func setBackoff(f BackoffFunc) applyOptionFunc {
	return func(opt *options) {
		opt.backoffFunc = f
	}
}
