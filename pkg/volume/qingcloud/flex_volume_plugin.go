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

	OptionFSType         = "kubernetes.io/fsType"
	OptionReadWrite      = "kubernetes.io/readwrite"
	OptionPVorVolumeName = "kubernetes.io/pvOrVolumeName"
	OptionVolumeID       = "volumeID"

	DefaultFSType  = "ext4"
	FlexDriverName = "qingcloud/flex-volume"

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
	volumeID, _ := options[OptionVolumeID].(string)
	pvOrVolumeName, _ := options[OptionPVorVolumeName].(string)
	// flexVolumeDriver GetVolumeName is not yet supported,  so PVorVolumeName is pvName, and store pvName to volumeName
	if !isVolumeID(pvOrVolumeName) {
		err := p.manager.UpdateVolume(volumeID, pvOrVolumeName)
		if err != nil {
			return flex.NewVolumeError("Error updating volume (%s) name to (%s) : %s", volumeID, pvOrVolumeName, err.Error())
		}
	}
	// VolumeManager.AttachVolume checks if disk is already attached to node and
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

func (p *flexVolumePlugin) Detach(pvOrVolumeName string, node string) flex.VolumeResult {
	var volumeID string
	var err error
	if !isVolumeID(pvOrVolumeName) {
		volumeID, err = p.manager.GetVolumeIDByName(pvOrVolumeName)
		if err != nil {
			return flex.NewVolumeError("Error GetVolumeIDByName (%s) : %s", pvOrVolumeName, err.Error())
		}
	} else {
		volumeID = pvOrVolumeName
	}

	if err = p.manager.DetachVolume(volumeID, node); err != nil {
		return flex.NewVolumeError("Error detaching volumeID %q: %v", volumeID, err)
	}
	return flex.NewVolumeSuccess()
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
		return flex.NewVolumeError("FormatAndMount device (%s) dir (%s) ", device, dir, err.Error())
	}
	return flex.NewVolumeSuccess()
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
	volumeID, _ := options[OptionVolumeID].(string)

	if device == "" {
		return flex.NewVolumeError("WaitForAttach failed for  Volume %q: device is empty.", volumeID)
	}

	ticker := time.NewTicker(checkSleepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			glog.V(4).Infof("Checking  volume %q is attached.", volumeID)
			exists, err := volumeutil.PathExists(device)
			if err != nil {
				// Log error, if any, and continue checking periodically.
				glog.Errorf("Error verifying  volume (%q) is attached: %v", volumeID, err)
			} else if exists {
				// A device path has successfully been created for the PD
				glog.Infof("Successfully found attached  volume %q.", volumeID)
				return flex.NewVolumeSuccess().WithDevicePath(device)
			}
		}
	}
}

func (*flexVolumePlugin) GetVolumeName(options flex.VolumeOptions) flex.VolumeResult {
	//TODO to implements this method when k8s 1.8 fix bug: https://github.com/kubernetes/kubernetes/issues/44737
	//and https://github.com/kubernetes/kubernetes/blob/f39c6087c2b2b473c37618d9cd054d918be0f77a/pkg/volume/flexvolume/plugin.go#L123
	// implements getvolumename call.
	// https://github.com/kubernetes/kubernetes/pull/46249
	return flex.NewVolumeNotSupported("getvolumename is not supported.")
}

func (p *flexVolumePlugin) IsAttached(options flex.VolumeOptions, node string) flex.VolumeResult {
	volumeID, _ := options[OptionVolumeID].(string)
	r, err := p.manager.VolumeIsAttached(volumeID, node)

	if err != nil {
		return flex.NewVolumeError(err.Error())
	}
	return flex.NewVolumeSuccess().WithAttached(r)
}
