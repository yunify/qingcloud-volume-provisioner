FROM busybox:1.27.1-glibc

COPY bin/qingcloud-volume-provisioner /qingcloud-volume-provisioner

RUN ln -s /qingcloud-volume-provisioner /bin/qingcloud-volume-provisioner