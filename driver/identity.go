package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/sirupsen/logrus"
)

func (d *Driver) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {

	resp := &csi.GetPluginInfoResponse{
		Name:          d.name,
		VendorVersion: Version,
	}
	d.logger.WithFields(logrus.Fields{
		"response": resp,
		"method":   "get_plugin_info",
	}).Info("get plugin info called")
	return resp, nil
}

func (d *Driver) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_ONLINE,
					},
				},
			},
		},
	}

	d.logger.WithFields(logrus.Fields{
		"response": resp,
		"method":   "get_plugin_capabilities",
	}).Info("get plugin capabitilies called")
	return resp, nil
}

func (d *Driver) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	d.logger.WithField("method", "probe").Info("probe called")
	d.readyMutex.Lock()
	defer d.readyMutex.Unlock()

	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{
			Value: d.ready,
		},
	}, nil
}
