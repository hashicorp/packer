package yandex

import (
	"context"

	ycsdk "github.com/yandex-cloud/go-sdk"
)

type Driver interface {
	DeleteImage(id string) error
	SDK() *ycsdk.SDK
	GetImage(imageID string) (*Image, error)
	GetImageFromFolder(ctx context.Context, folderID string, family string) (*Image, error)
	DeleteDisk(ctx context.Context, diskID string) error
	DeleteInstance(ctx context.Context, instanceID string) error
	DeleteSubnet(ctx context.Context, subnetID string) error
	DeleteNetwork(ctx context.Context, networkID string) error
}
