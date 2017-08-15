FROM busybox:1.27.1-glibc

COPY bin/qingcloud-volume-provisioner /qingcloud-volume-provisioner
COPY bin/qingcloud-flex-volume /qingcloud-flex-volume

RUN ln -s /qingcloud-volume-provisioner /bin/qingcloud-volume-provisioner
RUN ln -s /qingcloud-flex-volume /bin/qingcloud-flex-volume