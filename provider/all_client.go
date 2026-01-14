// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/minio/minio-go/v7"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

type AllClient struct {
	Minio      *minio.Client
	RustClient rustfs.RustfsAdmin
}
