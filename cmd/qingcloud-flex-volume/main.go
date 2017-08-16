package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	qclogger "github.com/yunify/qingcloud-sdk-go/logger"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/flex"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/qingcloud"
	"os"
	"path"
)

const (
	DriverDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
	LogDir    = "/var/log/qingcloud-flex-volume"
)

// printResult is a convenient method for printing result of volume operation, and return exit code.
func printResult(result flex.VolumeResult) int {
	fmt.Printf(result.ToJson())
	if result.Status == flex.StatusSuccess  {
		glog.Infof("ResponseSuccess: %#v", result.ToJson())
		return 0
	}
	if result.Status == flex.StatusNotSupported {
		glog.Infof("ResponseNotSupported : %#v", result.Error())
	}else {
		glog.Errorf("ResponseFailure : %#v", result.Error())
	}
	return 1
}

// ensureVolumeOptions decodes json or die
func ensureVolumeOptions(v string) (vo flex.VolumeOptions) {
	err := json.Unmarshal([]byte(v), &vo)
	if err != nil {
		panic(fmt.Errorf("Invalid json options: %s", v))
	}
	return
}

func installDriver(driverDir string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	vendor, driver := path.Split(qingcloud.FlexDriverName)
	vendor = path.Clean(vendor)
	driverTargetDir := path.Join(driverDir, fmt.Sprintf("%s~%s", vendor, driver))
	driverTargetFile := path.Join(driverTargetDir, driver)
	fmt.Printf("Install driver to %s \n", driverTargetFile)
	err = os.MkdirAll(driverTargetDir, 0644)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(driverTargetFile); !os.IsNotExist(err) {
		if err = os.Remove(driverTargetFile); err != nil {
			panic(err)
		}
	}
	err = os.Link(ex, driverTargetFile)
	if err != nil {
		panic(err)
	}
}

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer *GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

func handler(op string, args []string) flex.VolumeResult {
	volumePlugin, err := qingcloud.NewFlexVolumePlugin()

	if err != nil {
		return flex.NewVolumeError("Error init FlexVolumePlugin")
	}

	var ret flex.VolumeResult

	switch op {
	case "init":
		ret = volumePlugin.Init()
	case "attach":
		if len(args) < 2 {
			return flex.NewVolumeError("attach requires options in json format and a node name")
		}
		ret = volumePlugin.Attach(ensureVolumeOptions(args[0]), args[1])
	case "isattached":
		if len(args) < 2 {
			return flex.NewVolumeError("isattached requires options in json format and a node name")
		}
		ret = volumePlugin.Attach(ensureVolumeOptions(args[0]), args[1])
	case "detach":
		if len(args) < 2 {
			return flex.NewVolumeError("detach requires a device path and a node name")
		}
		ret = volumePlugin.Detach(args[0], args[1])
	case "mountdevice":
		if len(args) < 3 {
			return flex.NewVolumeError("mountdevice requires a mount path, a device path and mount options")
		}
		ret = volumePlugin.MountDevice(args[0], args[1], ensureVolumeOptions(args[2]))
	case "unmountdevice":
		if len(args) < 1 {
			return flex.NewVolumeError("unmountdevice requires a mount path")
		}
		ret = volumePlugin.UnmountDevice(args[0])
	case "waitforattach":
		if len(args) < 2 {
			return flex.NewVolumeError("waitforattach requires a device path and options in json format")
		}
		ret = volumePlugin.WaitForAttach(args[0], ensureVolumeOptions(args[1]))
	case "getvolumename":
		if len(args) < 1 {
			return flex.NewVolumeError("getvolumename requires options in json format")
		}
		ret = volumePlugin.GetVolumeName(ensureVolumeOptions(args[0]))
	default:
		ret = flex.NewVolumeNotSupported(op)
	}
	return ret
}

func mainFunc() int {

	var driverDir string = DriverDir
	install := flag.Bool("install", false, fmt.Sprintf("Install %s to %s", qingcloud.FlexDriverName, DriverDir))
	flag.StringVar(&driverDir, "driver_dir", DriverDir, "Driver dir to install.")

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = LogDir
	}
	// Prepare logs
	err := os.MkdirAll(logDir, 0750)
	if err != nil {
		panic(fmt.Sprintf("mkdir %s err: %s", logDir, err.Error()))
	}

	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Set("log_dir", logDir)
	flag.Set("stderrthreshold", "FATAL")
	flag.Set("v", "4")
	flag.Parse()
	glog.CopyStandardLogTo("INFO")
	defer glog.Flush()

	glog.Infof("Call %s driver, args: %#v", qingcloud.FlexDriverName, flag.Args())

	qclogger.SetOutput(&GlogWriter{})

	if *install {
		flag.VisitAll(func(f *flag.Flag) {
			glog.Infof("Flag: %s=%s", f.Name, f.Value)
		})
		installDriver(driverDir)
		return 0
	}
	var ret flex.VolumeResult
	args := flag.Args()
	if len(args) == 0 {
		ret = flex.NewVolumeError("Usage: %s init|attach|detach|mountdevice|unmountdevice|waitforattach|getvolumename|isattached", os.Args[0])
	} else {
		op := args[0]
		args = args[1:]
		ret = handler(op, args)
	}
	return printResult(ret)
}

func main() {
	os.Exit(mainFunc())
}
