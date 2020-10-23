// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package operation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

func TestOperation_Metadata_Unmarshal(t *testing.T) {
	expected := &wrappers.StringValue{Value: "metadata"}
	op := New(nil, &Proto{Metadata: marshalAny(t, expected)})
	actual, err := op.Metadata()
	require.NoError(t, err)
	assert.True(t, proto.Equal(expected, actual))
}

func TestOperation_Metadata_Nil(t *testing.T) {
	op := New(nil, &Proto{})
	actual, err := op.Metadata()
	require.NoError(t, err)
	assert.Nil(t, actual)
}

func TestOperation_Running(t *testing.T) {
	op := New(nil, &Proto{Done: false})
	assert.False(t, op.Done())
	assert.False(t, op.Ok())
	assert.False(t, op.Failed())
	assert.Nil(t, op.Error())
	assert.Nil(t, op.RawResponse())
	resp, err := op.Response()
	assert.Nil(t, resp)
	assert.NoError(t, err)
}

func TestOperation_Ok(t *testing.T) {
	resp := &wrappers.StringValue{Value: "response"}
	op := New(nil, &Proto{Done: true, Result: &operation.Operation_Response{Response: marshalAny(t, resp)}})
	assert.True(t, op.Done())
	assert.True(t, op.Ok())
	assert.False(t, op.Failed())
	assert.Nil(t, op.Error())

	assert.NotNil(t, op.RawResponse())
	actualResp, err := op.Response()
	assert.True(t, proto.Equal(resp, actualResp))
	assert.NoError(t, err)
}

func TestOperation_Fail(t *testing.T) {
	st := status.New(codes.Internal, "internal error")
	op := New(nil, &Proto{Done: true, Result: &operation.Operation_Error{Error: st.Proto()}})
	assert.True(t, op.Done())
	assert.False(t, op.Ok())
	assert.True(t, op.Failed())
	assert.True(t, proto.Equal(st.Proto(), op.ErrorStatus().Proto()), "should be equal to", st)

	assert.Nil(t, op.RawResponse())
}

func TestOperation_Poll_Ok(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	op := New(client, initialState)
	ctx := context.Background()
	newState := &Proto{Done: true, Result: &operation.Operation_Response{Response: marshalAny(t, &wrappers.StringValue{Value: "ok"})}}
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}).Return(newState, nil)

	err := op.Poll(ctx)
	require.NoError(t, err)
	assert.Equal(t, newState, op.Proto())

	client.AssertExpectations(t)
}

func TestOperation_Poll_Fail(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	op := New(client, initialState)
	ctx := context.Background()
	expectedErr := fmt.Errorf("test error")
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}).Return(nil, expectedErr)

	err := op.Poll(ctx)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, initialState, op.Proto())

	client.AssertExpectations(t)
}

func TestOperation_Cancel(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	op := New(client, initialState)
	ctx := context.Background()
	newState := &Proto{Id: id, Done: true}
	client.On("Cancel", ctx, &operation.CancelOperationRequest{OperationId: id}).Return(newState, nil)

	err := op.Cancel(ctx)
	require.NoError(t, err)
	assert.Equal(t, newState, op.Proto())

	client.AssertExpectations(t)
}

func TestOperation_Wait_Ok(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	op := New(client, initialState)
	ctx := context.Background()

	const expectedCalls = 3
	const interval = 10 * time.Millisecond
	var callNo int
	start := time.Now()
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}, mock.Anything).
		Return(func(ctx context.Context, in *operation.GetOperationRequest, _ ...grpc.CallOption) (*operation.Operation, error) {
			assert.True(t, time.Since(start) > time.Duration(callNo)*interval)
			callNo++
			if callNo >= expectedCalls {
				return &operation.Operation{Id: id, Done: true}, nil
			}
			return &operation.Operation{Id: id}, nil
		}).Times(expectedCalls)
	err := op.WaitInterval(ctx, interval)

	require.NoError(t, err)
	assert.True(t, op.Done())

	client.AssertExpectations(t)
}

func TestOperation_Wait_Failed(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	op := New(client, initialState)
	ctx := context.Background()
	st := status.New(codes.Internal, "internal error")
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}, mock.Anything).
		Return(&operation.Operation{Id: id, Done: true, Result: &operation.Operation_Error{Error: st.Proto()}}, nil)
	err := op.Wait(ctx)
	assert.Error(t, err)
	assert.True(t, proto.Equal(st.Proto(), status.Convert(err).Proto()), "error should be equal to", st, "but is", status.Convert(err))
	assert.Contains(t, err.Error(), id)
	assert.Contains(t, err.Error(), st.Err().Error())
}

func TestOperation_Wait_Interval(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	timer := &mockTimer{}
	op := New(client, initialState)
	op.newTimer = timer.New
	ctx := context.Background()

	const defaultInterval = 10 * time.Millisecond
	var callNo int
	expectedCalls := []string{
		"", "1", "", "12", "",
	}
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}, mock.Anything).
		Return(func(ctx context.Context, in *operation.GetOperationRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
			require.Equal(t, 1, len(opts))
			opt, ok := opts[0].(grpc.HeaderCallOption)
			require.True(t, ok)
			if val := expectedCalls[callNo]; val != "" {
				opt.HeaderAddr.Set(pollIntervalMetadataKey, expectedCalls[callNo])
			}
			callNo++
			if callNo >= len(expectedCalls) {
				return &operation.Operation{Id: id, Done: true}, nil
			}
			return &operation.Operation{Id: id}, nil
		}).Times(len(expectedCalls))

	timer.mock.On("Start", 1, defaultInterval).Once()
	timer.mock.On("Read", 1).Once()
	timer.mock.On("Start", 2, 1*time.Second).Once()
	timer.mock.On("Read", 2).Once()
	timer.mock.On("Start", 3, defaultInterval).Once()
	timer.mock.On("Read", 3).Once()
	timer.mock.On("Start", 4, 12*time.Second).Once()
	timer.mock.On("Read", 4).Once()

	err := op.WaitInterval(ctx, defaultInterval)

	require.NoError(t, err)
	assert.True(t, op.Done())

	client.AssertExpectations(t)
	timer.mock.AssertExpectations(t)
}

func TestOperation_RetryNotFound(t *testing.T) {
	const id = "aaa"
	client := &MockClient{}
	initialState := &Proto{Id: id}
	timer := &mockTimer{}
	ctx := context.Background()
	const defaultInterval = 10 * time.Millisecond
	wait := func() (bool, error) {
		op := New(client, initialState)
		op.newTimer = timer.New
		err := op.WaitInterval(ctx, defaultInterval)
		return op.Done(), err
	}
	timer.mock.On("Start", mock.Anything, defaultInterval)
	timer.mock.On("Read", mock.Anything)

	var callNo int
	var expectedNotFound = 3
	client.On("Get", ctx, &operation.GetOperationRequest{OperationId: id}, mock.Anything).
		Return(func(ctx context.Context, in *operation.GetOperationRequest, _ ...grpc.CallOption) (*operation.Operation, error) {
			if callNo < expectedNotFound {
				callNo++
				return nil, status.Errorf(codes.NotFound, "Not found yet %v", callNo)
			}
			return &operation.Operation{Id: id, Done: true}, nil
		})

	for _, v := range []struct {
		notFoundErrors int
		expectError    bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, false},
		{4, true},
		{5, true},
	} {
		t.Run(fmt.Sprintf("%v", v.notFoundErrors), func(t *testing.T) {
			callNo = 0
			expectedNotFound = v.notFoundErrors
			done, err := wait()
			if v.expectError {
				require.Error(t, err)
				assert.False(t, done)
			} else {
				require.NoError(t, err)
				assert.True(t, done)
			}
		})
	}
	client.AssertExpectations(t)
	timer.mock.AssertExpectations(t)
}
