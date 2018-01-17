// Copyright 2017 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

//go:generate go run gen_helper.go
//go:generate go fmt

package qingcloud

var (
	VERSION     string = "dev"
	GIT_SHA1    string = "dev+git"
	BUILD_LABEL string = "please use make to generate build files"
)
