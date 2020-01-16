package provider

import "context"

type VolumeProvider interface {
	GetVolumeType() VolumeType
	GetDetail() string
	CreateVolume(ctx context.Context, volId string, size int64) (string, error)
	DeleteVolume(ctx context.Context, volId string) error
	Attach(ctx context.Context, volId string) (string, error)
	Detach(ctx context.Context, volId string) error
	Resize(volId string, size int64) error
	GetMapDevice(volId string) (string, error)
	HasVolume(volId string) (bool, error)
}

type VolumeType string

const (
	CephType VolumeType = "ceph"
)
