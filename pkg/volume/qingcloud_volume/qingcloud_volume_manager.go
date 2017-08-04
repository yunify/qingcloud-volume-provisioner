package qingcloud_volume

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/volume"
	"strconv"
	"strings"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/qingcloud"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/api/core/v1"
)

// Abstract interface to PD operations.
type volumeManager interface {
	CreateVolume(provisioner *qingcloudVolumeProvisioner, options controller.VolumeOptions) (volumeID string, volumeSizeGB int, err error)
	DeleteVolume(deleter *qingcloudVolumeDeleter) error
}

type QingVolumeManager struct{}

func (manager *QingVolumeManager) DeleteVolume(d *qingcloudVolumeDeleter) error {
	glog.V(4).Infof("QingDiskUtil DeleteVolume called")

	var qcVolume qingcloud.Volumes
	var err error
	if qcVolume, err = getCloudProvider(d.qingcloudVolume.plugin.host.GetCloudProvider()); err != nil {
		return err
	}

	deleted, err := qcVolume.DeleteVolume(d.volumeID)
	if err != nil {
		glog.V(2).Infof("Error deleting qingcloud volume %s: %v", d.volumeID, err)
		return err
	}
	if deleted {
		glog.V(2).Infof("Successfully deleted qingcloud volume %s", d.volumeID)
	} else {
		glog.V(2).Infof("Successfully deleted qingcloud volume %s (actually already deleted)", d.volumeID)
	}
	return nil
}

// CreateVolume creates a qingcloud volume.
// Returns: volumeID, volumeSizeGB, error
func (manager *QingVolumeManager) CreateVolume(c *qingcloudVolumeProvisioner, options controller.VolumeOptions) (string, int, error) {
	glog.V(4).Infof("QingVolumeManager CreateVolume called, options: [%+v]", options)

	var qc *qingcloud.QingCloud
	var err error
	if qc, err = getCloudProvider(c.qingcloudVolume.plugin.host.GetCloudProvider()); err != nil {
		return "", 0, err
	}

	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestBytes := capacity.Value()
	// qingcloud works with gigabytes, convert to GiB with rounding up
	requestGB := int(volume.RoundUpSize(requestBytes, 1024*1024*1024))
	volumeOptions := &qingcloud.VolumeOptions{}

	// Apply Parameters (case-insensitive). We leave validation of
	// the values to the cloud provider.
	volType := sets.NewString("0", "2", "3")
	hasSetType := false
	for k, v := range options.Parameters {
		switch strings.ToLower(k) {
		case "type":
			if !volType.Has(v) {
				return "", 0, fmt.Errorf("invalid option '%q' for volume plugin %s, it only can be 0, 2, 3",
					k, c.plugin.GetPluginName())
			}
			volumeOptions.VolumeType, _ = strconv.Atoi(v)
			hasSetType = true
		default:
			return "", 0, fmt.Errorf("invalid option '%q' for volume plugin %s", k, c.plugin.GetPluginName())
		}
	}
	//auto set volume type by instance class.
	if !hasSetType {
		//TODO
		//selfInstance := qc.GetSelf()
		//if selfInstance.InstanceClass == nil || *selfInstance.InstanceClass == 0 {
		//	volumeOptions.VolumeType = 0
		//}else {
		//	volumeOptions.VolumeType = 3
		//}
		glog.V(2).Infof("Auto detected volume type: %v", volumeOptions.VolumeType)
	}

	//TODO refactor volume type define.
	switch volumeOptions.VolumeType {
	case 0:
		fallthrough
	case 3:
		// minimum 10GiB, maximum 500GiB
		if requestGB < 10 {
			requestGB = 10
		} else if requestGB > 1000 {
			return "", 0, fmt.Errorf("Can't request volume bigger than 1000GiB")
		}
		// must be a multiple of 10x
		if requestGB%10 != 0 {
			requestGB += 10 - requestGB%10
		}
	case 2:
		// minimum 100GiB, maximum 5000GiB
		if requestGB < 100 {
			requestGB = 100
		} else if requestGB > 50000 {
			return "", 0, fmt.Errorf("Can't request volume bigger than 5000GiB")
		}
		// must be a multiple of 50x
		if requestGB%50 != 0 {
			requestGB += 50 - requestGB%50
		}
	}

	volumeOptions.CapacityGB = requestGB

	// TODO: implement PVC.Selector parsing
	if options.PVC.Spec.Selector != nil {
		return "", 0, fmt.Errorf("claim.Spec.Selector is not supported for dynamic provisioning on qingcloud")
	}
	volumeOptions.VolumeName = fmt.Sprintf("k8s-%s-%s", options.PVC.Name, options.PVName)
	volumeID, err := qc.CreateVolume(volumeOptions)
	if err != nil {
		glog.V(2).Infof("Error creating qingcloud volume: %v", err)
		return "", 0, err
	}
	glog.V(2).Infof("Successfully created qingcloud volume %s", volumeID)

	return volumeID, int(requestGB), nil
}
