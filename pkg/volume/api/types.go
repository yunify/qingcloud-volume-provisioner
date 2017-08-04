package api

// Represents a Persistent Volume resource in qingcloud.
//
// A qingcloud volume must exist before mounting to a container. The volume
// must also be in the same qingcloud zone as the kubelet. A qingcloud volume
// can only be mounted as read/write once. qingcloud volumes support
// ownership management and SELinux relabeling.
type QingCloudStoreVolumeSource struct {
	// Unique id of the persistent volume resource. Used to identify the volume in qingcloud
	VolumeID string `json:"volumeID"`
	// Filesystem type to mount.
	// Must be a filesystem type supported by the host operating system.
	// Ex. "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if unspecified.
	// +optional
	FSType string `json:"fsType,omitempty"`
	// Optional: Defaults to false (read/write). ReadOnly here will force
	// the ReadOnly setting in VolumeMounts.
	ReadOnly bool `json:"readOnly,omitempty"`
}
