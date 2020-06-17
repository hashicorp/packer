// Code generated by sdkgen. DO NOT EDIT.

//nolint
package mysql

import (
	"context"

	"google.golang.org/grpc"

	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

//revive:disable

// BackupServiceClient is a mysql.BackupServiceClient with
// lazy GRPC connection initialization.
type BackupServiceClient struct {
	getConn func(ctx context.Context) (*grpc.ClientConn, error)
}

// Get implements mysql.BackupServiceClient
func (c *BackupServiceClient) Get(ctx context.Context, in *mysql.GetBackupRequest, opts ...grpc.CallOption) (*mysql.Backup, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return mysql.NewBackupServiceClient(conn).Get(ctx, in, opts...)
}

// List implements mysql.BackupServiceClient
func (c *BackupServiceClient) List(ctx context.Context, in *mysql.ListBackupsRequest, opts ...grpc.CallOption) (*mysql.ListBackupsResponse, error) {
	conn, err := c.getConn(ctx)
	if err != nil {
		return nil, err
	}
	return mysql.NewBackupServiceClient(conn).List(ctx, in, opts...)
}

type BackupIterator struct {
	ctx  context.Context
	opts []grpc.CallOption

	err     error
	started bool

	client  *BackupServiceClient
	request *mysql.ListBackupsRequest

	items []*mysql.Backup
}

func (c *BackupServiceClient) BackupIterator(ctx context.Context, folderId string, opts ...grpc.CallOption) *BackupIterator {
	return &BackupIterator{
		ctx:    ctx,
		opts:   opts,
		client: c,
		request: &mysql.ListBackupsRequest{
			FolderId: folderId,
			PageSize: 1000,
		},
	}
}

func (it *BackupIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if len(it.items) > 1 {
		it.items[0] = nil
		it.items = it.items[1:]
		return true
	}
	it.items = nil // consume last item, if any

	if it.started && it.request.PageToken == "" {
		return false
	}
	it.started = true

	response, err := it.client.List(it.ctx, it.request, it.opts...)
	it.err = err
	if err != nil {
		return false
	}

	it.items = response.Backups
	it.request.PageToken = response.NextPageToken
	return len(it.items) > 0
}

func (it *BackupIterator) Value() *mysql.Backup {
	if len(it.items) == 0 {
		panic("calling Value on empty iterator")
	}
	return it.items[0]
}

func (it *BackupIterator) Error() error {
	return it.err
}
