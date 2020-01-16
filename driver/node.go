package driver

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path"
)

func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Staging Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume Capability must be provided")
	}

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": req.StagingTargetPath,
		"method":              "node_stage_volume",
		"hostname":            d.hostname,
	})
	log.Info("node stage volume called")

	source := path.Base(req.VolumeId)
	target := req.StagingTargetPath

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	log = d.logger.WithFields(logrus.Fields{
		"volume_context":  req.VolumeContext,
		"publish_context": req.PublishContext,
		"source":          source,
		"target":          target,
		"vol_id":          req.VolumeId,
		"mount_options":   options,
	})

	deviceName, err := d.provider.Attach(ctx, req.VolumeId)
	if err != nil {
		return nil, err
	}

	log.Infof("start import ZPool, device: %s", deviceName)

	err = d.zfsUtil.ImportZPool(source)

	if err != nil {
		if err == NoSuchPoolErr {

			log.Infof("start create ZPool, source: %s", source)

			err = d.zfsUtil.CreateZPool(source, []string{deviceName})

			if err != nil {
				errMsg := fmt.Sprintf("Create zpool failed, %v", err)
				log.Error(errMsg)
				return nil, status.Errorf(codes.Internal, errMsg)
			}

		} else {
			errMsg := fmt.Sprintf("Import zpool failed, %v", err)
			log.Error(errMsg)
			return nil, status.Errorf(codes.Internal, errMsg)
		}
	}

	log.Info("mounting the volume for staging")

	zfsDirectory := path.Join(source, "globalmount")

	err = d.zfsUtil.CreateZFS(zfsDirectory)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = d.zfsUtil.Mount(path.Join("/", zfsDirectory), target, []string{"bind"}...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Info("formatting and mounting stage volume is finished")
	return &csi.NodeStageVolumeResponse{}, nil
}

func (d *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume Staging Target Path must be provided")
	}

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": req.StagingTargetPath,
		"method":              "node_unstage_volume",
	})
	log.Info("node unstage volume called")

	log.Info("start umount staging_target_path")

	err := d.zfsUtil.Unmount(req.StagingTargetPath)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Info("start export ZPool")

	source := path.Base(req.VolumeId)

	err = d.zfsUtil.ExportZPool(source)

	if err != nil && err != NoSuchPoolErr {
		return nil, status.Errorf(codes.Internal, "Export zpool failed: %v", err)
	}

	err = d.provider.Detach(ctx, req.VolumeId)

	if err != nil {
		return nil, err
	}

	log.Info("export ZPool is finished")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Staging Target Path must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume Capability must be provided")
	}

	source := req.StagingTargetPath

	target := req.TargetPath

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":           req.VolumeId,
		"staging_target_path": source,
		"target_path":         target,
		"method":              "node_publish_volume",
		"fstype":              req.VolumeCapability.GetMount().FsType,
	})
	log.Info("node publish volume called")

	err := d.zfsUtil.Mount(source, target, []string{"bind"}...)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Infof("bind mounting the volume is finished")
	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Volume ID must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Target Path must be provided")
	}

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":   req.VolumeId,
		"target_path": req.TargetPath,
		"method":      "node_unpublish_volume",
	})
	log.Info("node unpublish volume called")

	err := d.zfsUtil.Unmount(req.TargetPath)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Info("unmounting volume is finished")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume ID must be provided")
	}

	volumePath := req.VolumePath
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume Path must be provided")
	}

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":   req.VolumeId,
		"volume_path": volumePath,
		"method":      "node_get_volume_stats",
	})
	log.Info("node get volume stats called")

	found, err := d.provider.HasVolume(req.VolumeId)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "volume %q does not exist", req.VolumeId)
	}

	stats, err := d.zfsUtil.GetStatistics(volumePath)
	if err != nil {

		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound,
				"volume path %q not found, err: %s", volumePath, err.Error())
		}

		return nil, status.Errorf(codes.Internal,
			"failed to retrieve capacity statistics for volume path %q: %s", volumePath, err)
	}

	log.WithFields(logrus.Fields{
		"bytes_available":  stats.availableBytes,
		"bytes_total":      stats.totalBytes,
		"bytes_used":       stats.usedBytes,
		"inodes_available": stats.availableInodes,
		"inodes_total":     stats.totalInodes,
		"inodes_used":      stats.usedInodes,
	}).Info("node capacity statistics retrieved")

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			&csi.VolumeUsage{
				Available: stats.availableBytes,
				Total:     stats.totalBytes,
				Used:      stats.usedBytes,
				Unit:      csi.VolumeUsage_BYTES,
			},
			&csi.VolumeUsage{
				Available: stats.availableInodes,
				Total:     stats.totalInodes,
				Used:      stats.usedInodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

func (d *Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeExpandVolume volume ID not provided")
	}

	volumePath := req.GetVolumePath()
	if len(volumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeExpandVolume volume path not provided")
	}

	log := d.logger.WithFields(logrus.Fields{
		"volume_id":   req.VolumeId,
		"volume_path": req.VolumePath,
		"method":      "node_expand_volume",
	})
	log.Info("node expand volume called")

	resizeBytes, err := extractStorage(req.GetCapacityRange())
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "ControllerExpandVolume invalid capacity range: %v", err)
	}

	log.Infof("resize to %d", resizeBytes)

	err = d.provider.Resize(req.VolumeId, resizeBytes/miB)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Ceph resize %s failed, err: %s", req.VolumeId, err.Error())
	}

	deviceName, err := d.provider.GetMapDevice(req.VolumeId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Get map device for %s failed, err: %s",
			req.VolumeId, err.Error())
	}

	err = d.zfsUtil.Resize(req.VolumeId, deviceName)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Zpool Resize %s failed, err: %s", req.VolumeId, err.Error())
	}

	log.Info("volume was resized")
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (d *Driver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscaps := []*csi.NodeServiceCapability{
		&csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				},
			},
		},
		&csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
				},
			},
		},
		&csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
				},
			},
		},
	}

	d.logger.WithFields(logrus.Fields{
		"node_capabilities": nscaps,
		"method":            "node_get_capabilities",
	}).Info("node get capabilities called")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nscaps,
	}, nil
}

func (d *Driver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	d.logger.WithField("method", "node_get_info").Info("node get info called")
	return &csi.NodeGetInfoResponse{
		NodeId:            d.hostname,
		MaxVolumesPerNode: d.maxVolumesPerNode,
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{},
		},
	}, nil
}
