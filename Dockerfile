FROM alpine:edge AS build
RUN apk update
RUN apk upgrade
RUN apk add go gcc g++ make git linux-headers bash
WORKDIR /app
ENV GOPATH /app
ADD . /app/src/github.com/yunify/qingcloud-volume-provisioner
RUN cd /app/src/github.com/yunify/qingcloud-volume-provisioner && rm -rf bin/ && make

FROM alpine:latest
MAINTAINER calvinyu <calvinyu@yunify.com>

COPY --from=build /app/src/github.com/yunify/qingcloud-volume-provisioner/bin/qingcloud-volume-provisioner /bin/qingcloud-volume-provisioner
COPY --from=build /app/src/github.com/yunify/qingcloud-volume-provisioner/bin/qingcloud-flex-volume /bin/qingcloud-flex-volume
ENV PATH "/bin/qingcloud-volume-provisioner:/bin/qingcloud-flex-volume:$PATH"