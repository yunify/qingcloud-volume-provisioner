// Copyright 2017 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package qingcloud

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	qcconfig "github.com/yunify/qingcloud-sdk-go/config"
	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	waitInterval         = 10 * time.Second
	operationWaitTimeout = 180 * time.Second
)

func autoDetectedVolumeType(qcConfig *qcconfig.Config) (VolumeType, error) {
	qcService, err := qcservice.Init(qcConfig)
	if err != nil {
		return DefaultVolumeType, err
	}
	instanceService, err := qcService.Instance(qcConfig.Zone)
	if err != nil {
		return DefaultVolumeType, err
	}

	volumeType := DefaultVolumeType
	host, err := getHostname()
	if err == nil {
		ins, err := getInstanceByID(host, instanceService)
		if err == nil {
			if ins != nil {
				if ins.InstanceClass == nil || *ins.InstanceClass == 0 {
					volumeType = VolumeTypeHP
				} else {
					volumeType = VolumeTypeSHP
				}
				glog.V(2).Infof("Auto detected volume type: %v", volumeType)
			}
		} else {
			glog.Errorf("Get self instance fail, id: %s, err: %s", host, err.Error())
		}
	} else {
		glog.Errorf("Get Hostname fail, id: %s, err: %s", host, err.Error())
	}
	return volumeType, nil
}

func getHostname() (string, error) {
	content, err := ioutil.ReadFile("/etc/qingcloud/instance-id")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// getInstanceByID get instance.Instance by instanceId
func getInstanceByID(instanceID string, instanceService *qcservice.InstanceService) (*qcservice.Instance, error) {
	status := []*string{qcservice.String("running")}
	verbose := qcservice.Int(1)
	output, err := instanceService.DescribeInstances(&qcservice.DescribeInstancesInput{
		Instances:     []*string{&instanceID},
		Status:        status,
		Verbose:       verbose,
		IsClusterNode: qcservice.Int(1),
	})
	if err != nil {
		return nil, err
	}
	if len(output.InstanceSet) == 0 {
		return nil, nil
	}

	return output.InstanceSet[0], nil
}

//RoundUpVolumeCapacity return volume size in GiB
func RoundUpVolumeCapacity(capacity resource.Quantity, volumeType VolumeType) (int, error) {
	requestBytes := capacity.Value()
	// qingcloud works with gigabytes, convert to GiB with rounding up
	requestGB := int(volume.RoundUpSize(requestBytes, 1024*1024*1024))

	switch volumeType {
	case VolumeTypeHP:
		fallthrough
	case VolumeTypeSHP:
		// minimum 10GiB, maximum 1000GiB
		if requestGB < 10 {
			requestGB = 10
		} else if requestGB > 1000 {
			return 0, fmt.Errorf("Can't request volume bigger than 1000GiB")
		}
		// must be a multiple of 10x
		if requestGB%10 != 0 {
			requestGB += 10 - requestGB%10
		}
	case VolumeTypeHC:
		// minimum 100GiB, maximum 5000GiB
		if requestGB < 100 {
			requestGB = 100
		} else if requestGB > 5000 {
			return 0, fmt.Errorf("Can't request volume bigger than 5000GiB")
		}
		// must be a multiple of 50x
		if requestGB%50 != 0 {
			requestGB += 50 - requestGB%50
		}
	}
	return requestGB, nil
}

func isVolumeID(value string) bool {
	return strings.HasPrefix(value, "vol-")
}

func getVolumeName(path string) string {
	return filepath.Base(path)
}
