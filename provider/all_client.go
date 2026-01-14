package provider

import (
	"github.com/minio/minio-go/v7"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

type AllClient struct {
	Minio      *minio.Client
	RustClient rustfs.RustfsAdmin
}
