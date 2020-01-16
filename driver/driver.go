package driver

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/liuminjian/csi-zfs/provider"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
)

const (
	DefaultDriverName = "duni.csi.zfs.com"
	Version           = "0.0.1"
)

type Driver struct {
	name                  string
	publishInfoVolumeName string
	endpoint              string
	hostname              string
	srv                   *grpc.Server
	httpSrv               http.Server
	provider              provider.VolumeProvider
	readyMutex            sync.Mutex
	ready                 bool
	logger                *logrus.Entry
	zfsUtil               ZFSInterface
	maxVolumesPerNode     int64
	//记录已添加的存储和大小
	volumes map[string]int64
}

func NewDriver(endpoint string, driverName string, hostname string, maxVolumesPerNode int64,
	p provider.VolumeProvider, zfsUtil ZFSInterface) (*Driver, error) {

	log := logrus.New().WithFields(logrus.Fields{
		"volumeType": p.GetVolumeType(),
		"detail":     p.GetDetail(),
	})

	return &Driver{
		name:                  driverName,
		publishInfoVolumeName: driverName + "/volume-name",
		endpoint:              endpoint,
		hostname:              hostname,
		logger:                log,
		provider:              p,
		zfsUtil:               zfsUtil,
		maxVolumesPerNode:     maxVolumesPerNode,
		volumes:               make(map[string]int64),
	}, nil
}

func (d *Driver) Run(ctx context.Context) error {

	u, err := url.Parse(d.endpoint)

	if err != nil {
		return fmt.Errorf("unable to parse address: %v", err)
	}

	grpcAddr := path.Join(u.Host, filepath.FromSlash(u.Path))

	if u.Host == "" {
		grpcAddr = filepath.FromSlash(u.Path)
	}

	if u.Scheme != "unix" {
		return fmt.Errorf("currently only unix domain socket supported, have :%s", u.Scheme)
	}

	log := d.logger.WithFields(logrus.Fields{
		"socket": grpcAddr,
	})

	log.Info("removing socket")

	if err := os.Remove(grpcAddr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unix domain socket file: %s, error: %v", grpcAddr, err)
	}

	grpcListener, err := net.Listen(u.Scheme, grpcAddr)

	if err != nil {
		return fmt.Errorf("failed to listen:%v", err)
	}

	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			d.logger.WithError(err).WithField("method", info.FullMethod).Errorf("method failed")
		}
		return resp, err
	}
	d.ready = true
	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))

	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	log.Info("starting server")

	var eg errgroup.Group

	eg.Go(func() error {
		go func() {
			<-ctx.Done()
			d.logger.Info("grpc server stopped")
			d.srv.GracefulStop()
		}()
		return d.srv.Serve(grpcListener)
	})

	return eg.Wait()
}
