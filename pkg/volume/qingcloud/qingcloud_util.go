package qingcloud

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
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