package main

import (
	"context"
	"flag"
	"github.com/liuminjian/csi-zfs/driver"
	"github.com/liuminjian/csi-zfs/provider"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		endpoint = flag.String("endpoint", "unix:///var/lib/kubelet/plugins/"+
			driver.DefaultDriverName+"/csi.sock", "CSI endpoint")
		driverName        = flag.String("driver-name", driver.DefaultDriverName, "Name for the driver")
		volumeType        = flag.String("volume-type", "ceph", "block device type")
		timeout           = flag.Int("timeout", 60, "api timeout value")
		poolName          = flag.String("poolName", "test", "ceph pool Name")
		maxVolumesPerNode = flag.Int64("maxVolumesPerNode", 10000, "maxVolumesPerNode")
	)
	flag.Parse()

	var volumeSource provider.VolumeProvider

	if provider.VolumeType(*volumeType) == provider.CephType {
		volumeSource = provider.NewCephProvider(*timeout, *poolName)
	} else {
		log.Fatalln("volume-type is not right.")
	}

	hostname, _ := os.Hostname()

	drv, err := driver.NewDriver(*endpoint, *driverName, hostname, *maxVolumesPerNode, volumeSource,
		driver.NewZFSUtil())
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	if err := drv.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
