# qingcloud-volume-provisioner
QingCloud Volume External Storage Plugins for Kubernetes Provisoners

[![Build Status](https://travis-ci.org/yunify/qingcloud-volume-provisioner.svg?branch=master)](https://travis-ci.org/yunify/qingcloud-volume-provisioner)

English|[中文](README_zh.md)

**qingcloud-volume-provisioner** is a volume plugin deployed on QingCloud. This plugin will handle the volume operations requested from Kubernetes API server. Support IaaS :[QingCloud](http://qingcloud.com).

### Usage
1. Download the volume plugin file [qingcloud-flex-volume.tar.gz](https://pek3a.qingstor.com/k8s-qingcloud/k8s/qingcloud/volume/v1.1/qingcloud-flex-volume.tar.gz)  
1. Extract the package and grant the extracted file with excuting access:  
chmod +x *  
1. Run command as this:  
./qingcloud-flex-volume  --install=true  
1. Go to QingCloud console and create an API [access key](https://console.qingcloud.com/access_keys/)  
1. Create the config file /etc/qingcloud/client.yaml, which is used to access QingCloud IaaS resource, example as below:  
**<font color=red>qy_access_key_id: "your access key"  
qy_secret_access_key: "your secret key"  
zone: "your zone"</font>**  
log_level: warn  
connection_retries: 1  
connection_timeout: 5  
host: "api.ks.qingcloud.com"  
port: 80  
protocol: "http"  
1. Modify kubelet config file and config volume plugin, example as below:  
KUBELET_EXTRA_ARGS="--node-labels=role={{getv "/host/role"}},node_id={{getv "/host/node_id"}} --max-pods 60 --feature-gates=AllAlpha=true,DynamicKubeletConfig=false,RotateKubeletServerCertificate=false,RotateKubeletClientCertificate=false --root-dir=/data/var/lib/kubelet --cert-dir=/data/var/run/kubernetes **--enable-controller-attach-detach=true --volume-plugin-dir=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/**", please make sure **KUBELET_EXTRA_ARGS** is added into your kubelet.service file  
1. Download and deploy volume plugin pod by this config file [qingcloud-volume-provisioner.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/manifests/qingcloud-volume-provisioner.yaml), **please modify image version to v1.1**  
1. Download and deploy config file for QingCloud storage class as addons: [qingcloud-storage-class-capacity.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/addons/qingcloud/qingcloud-storage-class-capacity.yaml) and [qingcloud-storage-class.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/addons/qingcloud/qingcloud-storage-class.yaml)  
1. create logrotatte config file for this volume plugin /etc/logrotate.d/flex-volume  
/var/log/qingcloud-flex-volume/* {  
    rotate 1  
    copytruncate  
    missingok  
    notifempty  
    compress  
    maxsize 10M  
    daily  
    create 0644 root root  
}


