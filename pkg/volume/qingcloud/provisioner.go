package qingcloud

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	//"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"k8s.io/apimachinery/pkg/util/sets"
	"strconv"
)

const (
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "qingcloud-volume-provisioner"

	// VolumeGidAnnotationKey is the key of the annotation on the PersistentVolume
	// object that specifies a supplemental GID.
	VolumeGidAnnotationKey = "pv.beta.kubernetes.io/gid"

	// A PV annotation for the identity of the flexProvisioner that provisioned it
	annProvisionerId = "Provisioner_Id"
)

func NewProvisioner(qcConfigPath string) (controller.Provisioner, error) {
	manager,err := newVolumeManager(qcConfigPath)
	if err != nil {
		return nil, err
	}
	return &volumeProvisioner{manager: manager}, nil
}


type volumeProvisioner struct {
	manager VolumeManager
}

func (c *volumeProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	glog.V(4).Infof("qingcloudVolumeProvisioner Provision called, options: [%+v]", options)

	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestBytes := capacity.Value()
	// qingcloud works with gigabytes, convert to GiB with rounding up
	requestGB := int(volume.RoundUpSize(requestBytes, 1024*1024*1024))
	volumeOptions := &VolumeOptions{}

	// Apply Parameters (case-insensitive). We leave validation of
	// the values to the cloud provider.
	volType := sets.NewString("0", "2", "3")
	hasSetType := false
	for k, v := range options.Parameters {
		switch strings.ToLower(k) {
		case "type":
			if !volType.Has(v) {
				return nil, fmt.Errorf("invalid option '%q' for qingcloud-volume-provisioner, it only can be 0, 2, 3",
					k)
			}
			volumeOptions.VolumeType, _ = strconv.Atoi(v)
			hasSetType = true
		default:
			return nil, fmt.Errorf("invalid option '%q' for qingcloud-volume-provisioner", k)
		}
	}

	//auto set volume type by instance class.
	if !hasSetType {
		volumeOptions.VolumeType = c.manager.GetDefaultVolumeType()
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
			return nil, fmt.Errorf("Can't request volume bigger than 1000GiB")
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
			return nil, fmt.Errorf("Can't request volume bigger than 5000GiB")
		}
		// must be a multiple of 50x
		if requestGB%50 != 0 {
			requestGB += 50 - requestGB%50
		}
	}

	volumeOptions.CapacityGB = requestGB

	// TODO: implement PVC.Selector parsing
	if options.PVC.Spec.Selector != nil {
		return nil, fmt.Errorf("claim.Spec.Selector is not supported for dynamic provisioning on qingcloud")
	}
	volumeOptions.VolumeName = fmt.Sprintf("k8s-%s-%s", options.PVC.Name, options.PVName)
	volumeID, err := c.manager.CreateVolume(volumeOptions)
	if err != nil {
		glog.V(2).Infof("Error creating qingcloud volume: %v", err)
		return nil, err
	}
	glog.V(2).Infof("Successfully created qingcloud volume %s", volumeID)

	sizeGB := int(requestGB)

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerId] = "qingcloud-volume-provisioner"

	flexVolumeConfig := make(map[string]string)
	flexVolumeConfig["volumeID"] = volumeID
	//for key, value := range volumeConfig {
	//	flexVolumeConfig[key] = fmt.Sprintf("%v", value)
	//}


	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(fmt.Sprintf("%dGi", sizeGB)),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:    "qingcloud/volume",
					FSType:    "",
					SecretRef: nil,
					ReadOnly:  false,
					Options:   flexVolumeConfig,
				},
			},
		},
	}

	return pv, nil
}

func (c *volumeProvisioner) Delete(volume *v1.PersistentVolume) error{
	if volume.Name == "" {
		return fmt.Errorf("volume name cannot be empty %#v", volume)
	}

	if volume.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {
		if volume.Spec.PersistentVolumeSource.FlexVolume == nil {
			return fmt.Errorf("volume [%s] not support by qingcloud-volume-provisioner", volume.Name)
		}
		volumeID := volume.Spec.PersistentVolumeSource.FlexVolume.Options["volumeID"]
		if volumeID == ""{
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