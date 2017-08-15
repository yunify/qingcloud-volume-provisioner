#!/usr/bin/env bash

mkdir -p bin
go build -o bin/qingcloud-volume-provisioner ./cmd/qingcloud-volume-provisioner
go build -o bin/qingcloud-flex-volume ./cmd/qingcloud-flex-volume