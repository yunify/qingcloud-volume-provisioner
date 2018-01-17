# Copyright 2018 Yunify Inc. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM golang:1.9.2-alpine3.6 as builder

WORKDIR /go/src/github.com/yunify/qingcloud-volume-provisioner
COPY . .

RUN go install ./cmd/...

FROM alpine:3.6

LABEL MAINTAINER="calvinyu <calvinyu@yunify.com>"

COPY --from=builder /go/bin/* /usr/local/bin/
