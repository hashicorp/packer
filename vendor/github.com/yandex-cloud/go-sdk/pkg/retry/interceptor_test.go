package retry

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/pkg/testutil"
)

type ZoneServerHandler interface {
	handle(ctx context.Context) error
}

type ZoneServer struct {
	handler ZoneServerHandler
}

func (s *ZoneServer) Register(ser *grpc.Server) {
	compute.RegisterZoneServiceServer(ser, s)
}

var defaultZone = compute.Zone{}

func (s *ZoneServer) Get(ctx context.Context, _a1 *compute.GetZoneRequest) (*compute.Zone, error) {
	return &defaultZone, s.handler.handle(ctx)
}

func (s *ZoneServer) List(_a0 context.Context, _a1 *compute.ListZonesRequest) (*compute.ListZonesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

type serviceObjects struct {
	s *grpc.Server
	l net.Listener
	c *grpc.ClientConn
}

func (so *serviceObjects) cleanUp(t *testing.T) {
	if so.s != nil {
		so.s.Stop()
	} else {
		if so.l != nil {
			err := so.l.Close()
			require.NoError(t, err)
		}
	}

	if so.c != nil {
		err := so.c.Close()
		require.NoError(t, err)
	}
}

type cleanUp func(t *testing.T)

func initTestService(t *testing.T, handler ZoneServerHandler, interceptor grpc.DialOption) (res compute.ZoneServiceClient, clean cleanUp) {
	so := serviceObjects{}
	clean = so.cleanUp
	defer func() {
		if res == nil {
			clean(t)
		}
	}()

	var err error
	so.l, err = net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	so.s = grpc.NewServer()
	z := &ZoneServer{handler}
	z.Register(so.s)

	result := make(chan error)
	go func() {
		result <- so.s.Serve(so.l)
	}()

	testutil.Eventually(t,
		func() bool {
			return IsGrpcEndpointReady(t, so.l.Addr().String())
		},
		testutil.PollTimeout(1*time.Minute),
		testutil.Message("Test server failed to start."),
	)

	so.c, err = grpc.Dial(so.l.Addr().String(), grpc.WithInsecure(), interceptor)
	require.NoError(t, err)
	res = compute.NewZoneServiceClient(so.c)
	return
}

func IsGrpcEndpointReady(t *testing.T, addr string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return false
	}
	defer func() {
		errCl := conn.Close()
		require.NoError(t, errCl)
	}()

	// Any service client will do, we are interested in any response but Unavailable
	client := compute.NewZoneServiceClient(conn)
	_, err = client.List(context.Background(), &compute.ListZonesRequest{})
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() != codes.Unavailable
}

type failFirstAttempts struct {
	countGetFailures int
}

func (f *failFirstAttempts) handle(ctx context.Context) error {
	if f.countGetFailures > 0 {
		f.countGetFailures--
		return status.Error(codes.Unavailable, "")
	}

	return nil
}

func TestFiveRetries(t *testing.T) {
	ser := failFirstAttempts{5}
	c, cleanUp := initTestService(t, &ser, WithDefaultInterceptor())
	defer cleanUp(t)

	res, err := c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"}, WithMax(4))
	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.Unavailable, errCode)

	ser.countGetFailures = 5
	res, err = c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"}, WithMax(5))
	require.Equal(t, &defaultZone, res)
	require.NoError(t, err)
}

type testRetriableCodes struct {
	codes []codes.Code
	idx   int
}

func (f *testRetriableCodes) handle(ctx context.Context) error {
	if f.idx >= len(f.codes) {
		return nil
	}

	ret := status.Error(f.codes[f.idx], "")
	f.idx++
	return ret
}

func (f *testRetriableCodes) resetState() {
	f.idx = 0
}

func TestRetriableCodes(t *testing.T) {
	trc := testRetriableCodes{[]codes.Code{codes.ResourceExhausted, codes.Unavailable, codes.DataLoss}, 0}
	i := grpc.WithUnaryInterceptor(Interceptor(WithCodes(trc.codes...)))
	c, cleanUp := initTestService(t, &trc, i)
	defer cleanUp(t)

	var err error
	var res *compute.Zone

	for retryQty := 0; retryQty < len(trc.codes); retryQty++ {
		res, err = c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"}, WithMax(retryQty))
		require.Nil(t, res)
		require.Error(t, err)
		errCode := status.Code(err)
		require.Equal(t, trc.codes[retryQty], errCode)
		trc.resetState()
	}

	res, err = c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"}, WithMax(len(trc.codes)))
	require.NoError(t, err)
	require.Equal(t, &defaultZone, res)
}

type timeoutChecker interface {
	check() bool
}

type alwaysUnavailable struct {
	queryCount int
	tChecker   timeoutChecker
	error      bool
	n          chan int
}

func (f *alwaysUnavailable) handle(ctx context.Context) error {
	const notifyCount = 100
	if f.tChecker != nil {
		if !f.tChecker.check() {
			f.error = true
		}
	}

	f.queryCount++

	if f.n != nil && f.queryCount == notifyCount {
		f.n <- f.queryCount
	}

	return status.Error(codes.Unavailable, "")
}

func TestInfiniteRetriesAndCancel(t *testing.T) {
	i := grpc.WithUnaryInterceptor(Interceptor(WithCodes(codes.Unavailable), WithMax(-1), WithAttemptHeader(true)))
	n := make(chan int)
	c, cleanUp := initTestService(t, &alwaysUnavailable{n: n}, i)
	defer cleanUp(t)

	result := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	var res *compute.Zone

	go func() {
		var getErr error
		res, getErr = c.Get(ctx, &compute.GetZoneRequest{ZoneId: "id"})
		result <- getErr
	}()

	<-n
	cancel()
	err := <-result
	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.Canceled, errCode)
}

func testInfiniteRetriesDeadlineAndBackoff(t *testing.T, tChecker timeoutChecker, callOpts ...grpc.CallOption) {
	const contextTimeout = 5 * time.Second

	i := grpc.WithUnaryInterceptor(Interceptor(WithCodes(codes.Unavailable), WithMax(-1), WithAttemptHeader(true)))
	s := alwaysUnavailable{tChecker: tChecker}
	c, cleanUp := initTestService(t, &s, i)
	defer cleanUp(t)

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	res, err := c.Get(ctx, &compute.GetZoneRequest{ZoneId: "id"}, callOpts...)

	require.False(t, s.error)
	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.DeadlineExceeded, errCode)
}

func TestInfiniteRetriesAndDeadline(t *testing.T) {
	testInfiniteRetriesDeadlineAndBackoff(t, nil, WithBackoff(nil))
}

func TestBackoffExponentialWithJitterDeadline(t *testing.T) {
	testInfiniteRetriesDeadlineAndBackoff(t, nil, WithBackoff(DefaultBackoff()))
}

type timeLinearBackoff struct {
	callTime       *time.Time
	minimalTimeout time.Duration
}

func (t *timeLinearBackoff) check() bool {
	ct := time.Now()

	if t.callTime == nil {
		t.callTime = &ct
		return true
	}

	d := ct.Sub(*t.callTime)
	return d >= t.minimalTimeout
}

func TestBackoffLinearWithJitterDeadline(t *testing.T) {
	mTo := 0.9 * 50.0 * float64(time.Millisecond)
	tChecker := timeLinearBackoff{minimalTimeout: time.Duration(mTo)}
	testInfiniteRetriesDeadlineAndBackoff(t, &tChecker, WithBackoff(DefaultLinearJitterBackoff()))
}

type neverReturnsInTime struct {
	shutdown chan struct{}
}

func (f *neverReturnsInTime) handle(ctx context.Context) error {
	_, withTimeout := ctx.Deadline()

	if withTimeout {
		<-f.shutdown
	}

	return status.Error(codes.Unavailable, "")
}

func TestPerCallTimeout(t *testing.T) {
	i := grpc.WithUnaryInterceptor(Interceptor(WithMax(10), WithAttemptHeader(true), WithPerCallTimeout(50*time.Millisecond)))
	shutdown := make(chan struct{})
	c, cleanUp := initTestService(t, &neverReturnsInTime{shutdown: shutdown}, i)
	defer cleanUp(t)

	res, err := c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"})

	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.DeadlineExceeded, errCode)
	close(shutdown)
}

type testHeaderTokenAndRetryCount struct {
	queryCount  int
	token       string
	tokenError  bool
	headerError bool
}

func (f *testHeaderTokenAndRetryCount) handle(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		token, tokenPresent := md["idempotency-key"]

		if !tokenPresent || len(token) != 1 {
			f.tokenError = true
		} else {
			// store token on first call, on consequent calls, check that token didn't change
			if f.queryCount == 0 {
				f.token = token[0]
			} else {
				if f.token != token[0] {
					f.tokenError = true
				}
			}
		}

		if f.queryCount > 0 {
			retryMeta, countPresent := md["x-retry-attempt"]
			expectedValue := strconv.Itoa(f.queryCount)

			if !countPresent || len(retryMeta) != 1 || retryMeta[0] != expectedValue {
				f.headerError = true
			}
		}
	} else {
		f.tokenError = true
		f.headerError = true
	}

	f.queryCount++
	return status.Error(codes.Unavailable, "")
}

func TestHeaderTokenAndRetryCount(t *testing.T) {
	i := grpc.WithUnaryInterceptor(Interceptor(WithCodes(codes.Unavailable), WithMax(100), WithAttemptHeader(true)))
	s := testHeaderTokenAndRetryCount{}
	c, cleanUp := initTestService(t, &s, i)
	defer cleanUp(t)

	res, err := c.Get(context.Background(), &compute.GetZoneRequest{ZoneId: "id"})

	require.False(t, s.headerError)
	require.False(t, s.tokenError)
	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.Unavailable, errCode)
}

type testTokenUnchanged struct {
	token        string
	tokenChanged bool
}

func (f *testTokenUnchanged) handle(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		token, tokenPresent := md["idempotency-key"]

		if !tokenPresent || len(token) != 1 || token[0] != f.token {
			f.tokenChanged = true
		}
	} else {
		f.tokenChanged = true
	}

	return status.Error(codes.Unavailable, "")
}

func TestIdempotencyTokenNotChanged(t *testing.T) {
	i := grpc.WithUnaryInterceptor(Interceptor(WithCodes(codes.Unavailable), WithMax(100), WithAttemptHeader(true)))
	s := testTokenUnchanged{token: uuid.New().String()}
	c, cleanUp := initTestService(t, &s, i)
	defer cleanUp(t)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "idempotency-key", s.token)
	res, err := c.Get(ctx, &compute.GetZoneRequest{ZoneId: "id"})

	require.False(t, s.tokenChanged)
	require.Nil(t, res)
	require.Error(t, err)
	errCode := status.Code(err)
	require.Equal(t, codes.Unavailable, errCode)
}
