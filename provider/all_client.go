package provider

import (
	"github.com/aminueza/terraform-provider-minio/minio"
	"github.com/weinmann-emt/terraform-provider-rustfs/pkg/rustfs"
)

type AllClient struct {
	minio.S3MinioClient
	RustClient rustfs.RustfsAdmin
}
