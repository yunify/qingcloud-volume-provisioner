package qingcloud

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	qcconfig "github.com/yunify/qingcloud-sdk-go/config"
	"github.com/golang/glog"
	"io/ioutil"
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


func autoDetectedVolumeType(qcConfig *qcconfig.Config) (int, error) {
	qcService, err := qcservice.Init(qcConfig)
	if err != nil {
		return nil, err
	}
	instanceService, err := qcService.Instance(qcConfig.Zone)
	if err != nil {
		return nil, err
	}

	var volumeType int = 0
	host, err := getHostname()
	if err == nil {
		ins, err := getInstanceByID(host, instanceService)
		if err == nil {
			if ins != nil {
				if ins.InstanceClass == nil || *ins.InstanceClass == 0 {
					volumeType = 0
				}else {
					volumeType = 3
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
