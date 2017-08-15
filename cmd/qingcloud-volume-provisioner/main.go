package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/qingcloud"
)

var (
	provisionerName = qingcloud.ProvisionerName
	master          = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig      = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	qcConfig	= flag.String("qingcloud_config", qingcloud.DefaultQingCloudConfigPath, "Absolute path to qingcloud client config yaml.")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	glog.Infof("Provisioner %s specified", provisionerName)

	// Create the client according to whether we are running in or out-of-cluster
	var config *rest.Config
	var err error
	if *master != "" || *kubeconfig != "" {
		glog.Infof("Either master or kubeconfig specified. building kube config from that..")
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		glog.Infof("Building kube configs for running in cluster...")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	provisioner, _ := qingcloud.NewProvisioner(*qcConfig)

	// Start the provision controller which will dynamically provision NFS PVs
	pc := controller.NewProvisionController(
		clientset,
		provisionerName,
		provisioner,
		serverVersion.GitVersion,
	)

	pc.Run(wait.NeverStop)
}