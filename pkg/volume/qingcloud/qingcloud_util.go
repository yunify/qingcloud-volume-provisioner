package qingcloud

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	qcconfig "github.com/yunify/qingcloud-sdk-go/config"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/resource"
	"fmt"
)

const (
	waitInterval         = 10 * time.Second
	operationWaitTimeout = 180 * time.Second
)

// Make sure qingcloud instance hostname or override-hostname (if provided) is equal to InstanceId
// Recommended to use override-hostname
func NodeNameToInstanceID(name types.NodeName) string {
	return string(name)
}

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
				}else {
					volumeType = VolumeTypeSHP
				}
				glog.V(2).Infof("Auto detected volume type: %v", volumeType)
			}
		}else{
			glog.Errorf("Get self instance fail, id: %s, err: %s", host, err.Error())
		}
	}else {
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
		Instances: []*string{&instanceID},
		Status:    status,
		Verbose:   verbose,
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

func fixVolumeCapacity(capacity *resource.Quantity, volumeType VolumeType) (*resource.Quantity, error) {
	// qingcloud works with gigabytes, convert to GiB with rounding up
	requestGB := int(capacity.ScaledValue(resource.Giga))

	switch volumeType {
	case VolumeTypeHP:
		fallthrough
	case VolumeTypeSHP:
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
	case VolumeTypeHC:
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
	return resource.NewScaledQuantity(int64(requestGB), resource.Giga), nil
}