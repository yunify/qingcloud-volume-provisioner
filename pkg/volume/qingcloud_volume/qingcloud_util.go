package qingcloud_volume

import (
	"k8s.io/kubernetes/pkg/cloudprovider"
	"time"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/qingcloud"
)

const (
	checkSleepDuration = time.Second
)

// Return cloud provider
func getCloudProvider(cloudProvider cloudprovider.Interface) (*qingcloud.QingCloud, error) {
	//TODO
	return nil, nil
}
