// Copyright 2017 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package qingcloud

import (
	"fmt"

	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeType int

const (
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "qingcloud-volume-provisioner"

	ProvisionerName = "qingcloud/volume-provisioner"

	// VolumeGidAnnotationKey is the key of the annotation on the PersistentVolume
	// object that specifies a supplemental GID.
	VolumeGidAnnotationKey = "pv.beta.kubernetes.io/gid"

	// A PV annotation for the identity of the flexProvisioner that provisioned it
	annProvisionerId = "Provisioner_Id"

	// fsType from annotation, will override the fstype parameter from storageclass
	annFsType = "kubernetes.io/fsType"
)

func NewProvisioner(qcConfigPath string) (controller.Provisioner, error) {
	manager, err := newVolumeManager(qcConfigPath)
	if err != nil {
		return nil, err
	}
	return &volumeProvisioner{manager: manager}, nil
}

type volumeProvisioner struct {
	manager VolumeManager
}

func (c *volumeProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	glog.V(4).Infof("Qingcloud volume Provisioner Provision called, options: [%+v]", options)

	// TODO: implement PVC.Selector parsing
	if options.PVC.Spec.Selector != nil {
		return nil, fmt.Errorf("claim.Spec.Selector is not supported for dynamic provisioning on qingcloud")
	}

	// Validate access modes
	found := false
	for _, mode := range options.PVC.Spec.AccessModes {
		if mode == v1.ReadWriteOnce {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Qingcloud volume only supports ReadWriteOnce mounts")
	}

	volumeOptions := &VolumeOptions{}

	hasSetType := false
	fsType := DefaultFSType
	for k, v := range options.Parameters {
		switch strings.ToLower(k) {
		case "type":
			if !supportVolumeTypes.Has(v) {
				return nil, fmt.Errorf("invalid option '%q' for qingcloud-volume-provisioner, it only can be 0, 2, 3",
					k)
			}
			volumeTypeInt, _ := strconv.Atoi(v)
			volumeOptions.VolumeType = VolumeType(volumeTypeInt)
			hasSetType = true
		case "fstype":
			glog.V(4).Infof("fstype %v is set by storageclass", v)
			fsType = v
		default:
			return nil, fmt.Errorf("invalid option '%q' for qingcloud-volume-provisioner", k)
		}
	}

	//auto set volume type by instance class.
	if !hasSetType {
		volumeOptions.VolumeType = c.manager.GetDefaultVolumeType()
	}

	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	sizeGB, err := RoundUpVolumeCapacity(capacity, volumeOptions.VolumeType)
	if err != nil {
		return nil, err
	}
	volumeOptions.CapacityGB = sizeGB

	//use pv name as volumeName
	volumeOptions.VolumeName = options.PVName
	volumeID, err := c.manager.CreateVolume(volumeOptions)
	if err != nil {
		glog.V(2).Infof("Error creating qingcloud volume: %v", err)
		return nil, err
	}
	glog.V(2).Infof("Successfully created qingcloud volume %s", volumeID)

	storageClassName := ""
	if options.PVC.Spec.StorageClassName != nil {
		storageClassName = *options.PVC.Spec.StorageClassName
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerId] = ProvisionerName

	flexVolumeConfig := make(map[string]string)
	flexVolumeConfig[OptionVolumeID] = volumeID

	annFsTypeVal, existed := options.PVC.ObjectMeta.Annotations[annFsType]

	if existed {
		glog.V(4).Infof("fstype %v is set in annatation and will use it to format volume", annFsTypeVal)
		fsType = annFsTypeVal
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(fmt.Sprintf("%dGi", sizeGB)),
			},
			StorageClassName: storageClassName,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:   FlexDriverName,
					FSType:   fsType,
					ReadOnly: false,
					Options:  flexVolumeConfig,
				},
			},
		},
	}

	return pv, nil
}

func (c *volumeProvisioner) Delete(volume *v1.PersistentVolume) error {
	if volume.Name == "" {
		return fmt.Errorf("volume name cannot be empty %#v", volume)
	}

	if volume.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {
		if volume.Spec.PersistentVolumeSource.FlexVolume == nil {
			return fmt.Errorf("volume [%s] not support by qingcloud-volume-provisioner", volume.Name)
		}
		volumeID := volume.Spec.PersistentVolumeSource.FlexVolume.Options["volumeID"]
		if volumeID == "" {
			return fmt.Errorf("Spec.PersistentVolumeSource.FlexVolume.Options[\"volumeID\"]  cannot be empty %#v", volume)
		}
		_, err := c.manager.DeleteVolume(volumeID)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
