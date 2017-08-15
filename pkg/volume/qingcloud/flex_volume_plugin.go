package qingcloud

import (
	"github.com/golang/glog"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/flex"
	"k8s.io/kubernetes/pkg/util/exec"
	"k8s.io/kubernetes/pkg/util/mount"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
	"os"
	"strings"
	"time"
)

const (
	checkSleepDuration = time.Second

	OptionFSType    = "kubernetes.io/fsType"
	OptionReadWrite = "kubernetes.io/readwrite"

	DefaultFSType = "ext4"
	DriverName    = "qingcloud/flex-volume"

	DefaultQingCloudConfigPath = "/etc/qingcloud/client.yaml"
)

type flexVolumePlugin struct {
	manager VolumeManager
}

func NewFlexVolumePlugin() (flex.VolumePlugin, error) {
	manager, err := newVolumeManager(DefaultQingCloudConfigPath)
	if err != nil {
		return nil, err
	}
	return &flexVolumePlugin{manager: manager}, nil
}

func (*flexVolumePlugin) Init() flex.VolumeResult {
	return flex.NewVolumeSuccess()
}

func (p *flexVolumePlugin) Attach(options flex.VolumeOptions, node string) flex.VolumeResult {
	volumeID, _ := options["volumeID"].(string)

	// qingcloud.AttachVolume checks if disk is already attached to node and
	// succeeds in that case, so no need to do that separately.
	devicePath, err := p.manager.AttachVolume(volumeID, node)
	if err != nil {
		//ignore already attached error
		if !strings.Contains(err.Error(), "have been already attached to instance") {
			glog.Errorf("Error attaching volume %q: %+v", volumeID, err)
			return flex.NewVolumeError("Error attaching volume %q to node %s: %+v", volumeID, node, err)
		}
	}
	return flex.NewVolumeSuccess().WithDevicePath(devicePath)
}

func (*flexVolumePlugin) Detach(device string, node string) flex.VolumeResult {
	panic("implement me")
}

func (*flexVolumePlugin) MountDevice(dir, device string, options flex.VolumeOptions) flex.VolumeResult {
	fstype, _ := options[OptionFSType].(string)
	if fstype == "" {
		fstype = DefaultFSType
	}
	readwrite, _ := options[OptionReadWrite].(string)
	flagstr, _ := options["flags"].(string)
	flags := []string{}
	if flagstr != "" {
		flags = strings.Split(flagstr, ",")
	}
	if readwrite != "" {
		flags = append(flags, readwrite)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0750); err != nil {
			return flex.NewVolumeError(err.Error())
		}
	}

	volumeMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Runner: exec.New()}
	err := volumeMounter.FormatAndMount(device, dir, fstype, flags)
	if err != nil {
		os.Remove(dir)
		return flex.NewVolumeError(err.Error())
	}
	return flex.NewVolumeSuccess().WithDevicePath(device).WithAttached(true)
}

func (*flexVolumePlugin) UnmountDevice(dir string) flex.VolumeResult {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Runner: exec.New()}
	err := mounter.Unmount(dir)
	if err != nil {
		return flex.NewVolumeError(err.Error())
	}
	return flex.NewVolumeSuccess()
}

func (*flexVolumePlugin) WaitForAttach(device string, options flex.VolumeOptions) flex.VolumeResult {
	volumeID, _ := options["volumeID"].(string)

	if device == "" {
		return flex.NewVolumeError("WaitForAttach failed for qingcloud Volume %q: device is empty.", volumeID)
	}

	ticker := time.NewTicker(checkSleepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			glog.V(5).Infof("Checking qingcloud volume %q is attached.", volumeID)
			exists, err := volumeutil.PathExists(device)
			if err != nil {
				// Log error, if any, and continue checking periodically.
				glog.Errorf("Error verifying qingcloud volume (%q) is attached: %v", volumeID, err)
			} else if exists {
				// A device path has successfully been created for the PD
				glog.Infof("Successfully found attached qingcloud volume %q.", volumeID)
				return flex.NewVolumeSuccess().WithDevicePath(device)
			}
		}
	}
}

func (*flexVolumePlugin) GetVolumeName(options flex.VolumeOptions) flex.VolumeResult {
	volumeID, _ := options["volumeID"].(string)
	if volumeID == "" {
		return flex.NewVolumeError("volumeID is required, options: %+v", options)
	}
	return flex.NewVolumeSuccess().WithVolumeName(volumeID)
}

func (p *flexVolumePlugin) IsAttached(options flex.VolumeOptions, node string) flex.VolumeResult {
	volumeID, _ := options["volumeID"].(string)
	r, err := p.manager.VolumeIsAttached(volumeID, node)

	if err != nil {
		return flex.NewVolumeError(err.Error())
	}
	return flex.NewVolumeSuccess().WithAttached(r)
}
