# qingcloud-volume-provisioner
QingCloud Volume External Storage Plugins for Kubernetes Provisoners

[![Build Status](https://travis-ci.org/yunify/qingcloud-volume-provisioner.svg?branch=master)](https://travis-ci.org/yunify/qingcloud-volume-provisioner)

English|[中文](README_zh.md)

**qingcloud-volume-provisioner** is a volume plugin deployed on QingCloud. This plugin will handle the volume operations requested from Kubernetes API server. Support IaaS :[QingCloud](http://qingcloud.com).

### Usage
1. Download the volume plugin binary file from and save it here:
1. Extract the package and grant the extracted file with excuting access
1. Run command as this: ./qingcloud-flex-volume  --install=true
1. Create config file /etc/qingcloud/client.yaml

