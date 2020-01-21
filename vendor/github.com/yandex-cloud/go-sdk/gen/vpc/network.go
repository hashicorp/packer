// Code generated by sdkgen. DO NOT EDIT.

//nolint
package vpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

//revive:disable

// NetworkServiceClient is a vpc.NetworkServiceClient with
// lazy GRPC connection initialization.
type NetworkServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

var _ vpc.NetworkServiceClient = &NetworkServiceClient{}

// Create implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) Create(ctx context.Context, in *vpc.CreateNetworkRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).Create(ctx, in, opts...)
}

// Delete implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) Delete(ctx context.Context, in *vpc.DeleteNetworkRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).Delete(ctx, in, opts...)
}

// Get implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) Get(ctx context.Context, in *vpc.GetNetworkRequest, opts ...grpc.CallOption) (*vpc.Network, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).Get(ctx, in, opts...)
}

// List implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) List(ctx context.Context, in *vpc.ListNetworksRequest, opts ...grpc.CallOption) (*vpc.ListNetworksResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).List(ctx, in, opts...)
}

// ListOperations implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) ListOperations(ctx context.Context, in *vpc.ListNetworkOperationsRequest, opts ...grpc.CallOption) (*vpc.ListNetworkOperationsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).ListOperations(ctx, in, opts...)
}

// ListSubnets implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) ListSubnets(ctx context.Context, in *vpc.ListNetworkSubnetsRequest, opts ...grpc.CallOption) (*vpc.ListNetworkSubnetsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).ListSubnets(ctx, in, opts...)
}

// Update implements vpc.NetworkServiceClient
func (c *NetworkServiceClient) Update(ctx context.Context, in *vpc.UpdateNetworkRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return vpc.NewNetworkServiceClient(conn).Update(ctx, in, opts...)
}
