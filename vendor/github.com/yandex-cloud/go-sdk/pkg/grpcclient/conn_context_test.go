// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package grpcclient

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-sdk/pkg/testutil"
)

func startServer(ctx context.Context, t *testing.T, l net.Listener) {
	s := grpc.NewServer()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()
	err := s.Serve(l)
	require.NoError(t, err)
}

func TestDialContextTimeout(t *testing.T) {
	connCtx := NewLazyConnContext(DialOptions(
		grpc.WithInsecure(),
		grpc.WithTimeout(time.Millisecond*1), // nolint
		grpc.WithBlock(),
	))
	const addr = "blablabla:1234"
	x, err := connCtx.GetConn(context.Background(), addr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), addr)
	assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
	assert.Nil(t, x)
}

func TestNewLazyClientConnContext(t *testing.T) {
	addresses := []string{
		":1488",
		":1499",
	}

	ctx, cancel := context.WithCancel(context.Background())
	for _, addr := range addresses {
		l, err := net.Listen("tcp", addr)
		require.NoError(t, err)
		go func() {
			startServer(ctx, t, l)
		}()
	}
	defer cancel()

	connCtx := NewLazyConnContext(DialOptions(grpc.WithInsecure()))

	wg := sync.WaitGroup{}
	const numClients = 50
	const numRequests = 100
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func() {
			defer wg.Done()
			for r := 0; r < numRequests; r++ {
				idx := rand.Intn(len(addresses))
				_, err := connCtx.GetConn(ctx, addresses[idx])
				if err != nil {
					assert.Equal(t, ErrConnContextClosed, err)
				}
			}
		}()
	}
	time.Sleep(2 * time.Millisecond)
	err := connCtx.Shutdown(ctx)
	require.NoError(t, err)

	testutil.Eventually(t, func() bool {
		wg.Wait()
		return true
	})
}
