package test

import (
	"context"
	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	"github.com/liuminjian/csi-zfs/driver"
	"github.com/liuminjian/csi-zfs/provider"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"path"
	"testing"
)

func TestDriverSuite(t *testing.T) {

	dir, _ := os.Getwd()
	endpoint := "unix://" + dir + "/csi.sock"
	drv, err := driver.NewDriver(endpoint, driver.DefaultDriverName, "hostname", 100,
		&fakeProvider{volumes: make(map[string]int64)}, &fakeZFSUtil{})
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var eg errgroup.Group
	eg.Go(func() error {
		return drv.Run(ctx)
	})

	targetPath := dir + "/csi-target"
	stagePath := dir + "/csi-staging"

	os.Remove(targetPath)
	os.Remove(stagePath)

	cfg := &sanity.Config{
		TargetPath:  dir + "/csi-target",
		StagingPath: dir + "/csi-staging",
		Address:     endpoint,
	}
	sanity.Test(t, cfg)

	cancel()
	if err := eg.Wait(); err != nil {
		t.Errorf("driver run failed: %s", err)
	}
}

type fakeZFSUtil struct {
}

func (f *fakeZFSUtil) ImportZPool(poolName string) error {
	return nil
}

func (f *fakeZFSUtil) CreateZPool(poolName string, devices []string) error {
	return nil
}

func (f *fakeZFSUtil) CreateZFS(pathName string) error {
	return nil
}

func (f *fakeZFSUtil) ExportZPool(poolName string) error {
	return nil
}

func (f *fakeZFSUtil) Mount(source string, target string, opts ...string) error {
	return nil
}

func (f *fakeZFSUtil) Unmount(target string) error {
	return nil
}

func (f *fakeZFSUtil) GetStatistics(volumePath string) (driver.VolumeStatistics, error) {
	_, err := os.Stat(path.Dir(volumePath))
	if err != nil {
		return driver.VolumeStatistics{}, err
	}
	return driver.VolumeStatistics{}, nil
}

func (f *fakeZFSUtil) Resize(poolName string, deviceName string) error {
	return nil
}

type fakeProvider struct {
	volumes map[string]int64
}

func (f *fakeProvider) HasVolume(volId string) (bool, error) {
	_, ok := f.volumes[volId]
	return ok, nil
}

func (f *fakeProvider) GetVolumeType() provider.VolumeType {
	return provider.VolumeType("fake")
}

func (f *fakeProvider) GetDetail() string {
	return "fake provider"
}

func (f *fakeProvider) CreateVolume(ctx context.Context, volId string, size int64) (string, error) {
	f.volumes[volId] = size
	return volId, nil
}

func (f *fakeProvider) DeleteVolume(ctx context.Context, volId string) error {
	return nil
}

func (f *fakeProvider) Attach(ctx context.Context, volId string) (string, error) {
	return volId, nil
}

func (f *fakeProvider) Detach(ctx context.Context, volId string) error {
	return nil
}

func (f *fakeProvider) Resize(volId string, size int64) error {
	return nil
}

func (f *fakeProvider) GetMapDevice(volId string) (string, error) {
	return "", nil
}
