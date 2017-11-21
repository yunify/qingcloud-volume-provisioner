# qingcloud-volume-provisioner
QingCloud Volume External Storage Plugins for Kubernetes Provisoners

[![Build Status](https://travis-ci.org/yunify/qingcloud-volume-provisioner.svg?branch=master)](https://travis-ci.org/yunify/qingcloud-volume-provisioner)

中文|[English](README.md)

**qingcloud-volume-provisioner** 是部署于Kubernetes之上用于访问青云存储服务的插件。通过这个插件可以基于Kubernetes自动完成对青云硬盘的创建，挂载，格式化，卸载和删除等操作。操作对象为青云IaaS平台上面的容量盘，性能盘以及高性能盘: [QingCloud](http://qingcloud.com).

### 使用说明
1. 下载存储插件的压缩文件 [qingcloud-flex-volume.tar.gz](https://pek3a.qingstor.com/k8s-qingcloud/k8s/qingcloud/volume/v1.1/qingcloud-flex-volume.tar.gz)  
1. 解压并赋予解压文件可执行权限：  
chmod +x *  
1. 执行如下命令：  
./qingcloud-flex-volume  --install=true  
1. 访问青云的控制台创建 API [access key](https://console.qingcloud.com/access_keys/)  
1. 在您k8s的各个节点创建用于访问青云IaaS资源的配置文件 /etc/qingcloud/client.yaml，参考如下示例将您上一步创建的access key以及k8s所在zone保存至config.yaml：  
**qy_access_key_id: "your access key"**  
**qy_secret_access_key: "your secret key"**  
**zone: "your zone"**   
log_level: warn  
connection_retries: 1  
connection_timeout: 5  
host: "api.ks.qingcloud.com"  
port: 80  
protocol: "http"  
1. 修改kubelet配置文件指向青云存储插件  
KUBELET_EXTRA_ARGS="--node-labels=role={{getv "/host/role"}},node_id={{getv "/host/node_id"}} --max-pods 60 --feature-gates=AllAlpha=true,DynamicKubeletConfig=false,RotateKubeletServerCertificate=false,RotateKubeletClientCertificate=false --root-dir=/data/var/lib/kubelet --cert-dir=/data/var/run/kubernetes **--enable-controller-attach-detach=true --volume-plugin-dir=/usr/libexec/kubernetes/kubelet-plugins/volume/exec/**", <font color=red>请确保 **KUBELET_EXTRA_ARGS** 已被添加到 kubelet.service file</font> 
1. 修改controller manager的配置文件，打开flex volume相关配置的选项，可以参考 [kube-controller-manager.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/manifests/kube-controller-manager.yaml
), 通过搜索关键字“flex”找到相关参考配置  
1. 下载存储插件的配置文件并部署于 k8s [qingcloud-volume-provisioner.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/manifests/qingcloud-volume-provisioner.yaml), **请修改镜像版本为 v1.1**  
1. 以addon方式下载有关青云存储分级的配置文件并部署到 k8s 之上： [qingcloud-storage-class-capacity.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/addons/qingcloud/qingcloud-storage-class-capacity.yaml) 和 [qingcloud-storage-class.yaml](https://github.com/QingCloudAppcenter/kubernetes/blob/master/k8s/addons/qingcloud/qingcloud-storage-class.yaml)  
1. 为存储插件创建如下的 logrotatte 配置文件 /etc/logrotate.d/flex-volume  
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

**注意：步骤1，2，3，5，6，10是针对所有k8s节点的**


