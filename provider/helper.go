package provider

import (
	"os"

	"github.com/aminueza/terraform-provider-minio/minio"
)

func generateMinioConfig(model RustfsProviderModel) (config *minio.S3MinioConfig, err error) {

	endpoint := os.Getenv("RUSTFS_ENDPOINT")
	if endpoint == "" {
		endpoint = model.Endpoint.String()
	}

	user := os.Getenv("RUSTFS_USER")
	if user == "" {
		user = model.AccessKey.String()
	}

	secret := os.Getenv("RUSTFS_SECRET")
	if secret == "" {
		secret = model.AccessSecret.String()
	}

	config = &minio.S3MinioConfig{
		S3HostPort:      endpoint,
		S3Region:        "us-east-1",
		S3UserAccess:    user,
		S3UserSecret:    secret,
		S3SessionToken:  "",
		S3APISignature:  "v4",
		S3SSL:           model.Ssl.ValueBool(),
		S3SSLCACertFile: "",
		S3SSLCertFile:   "",
		S3SSLKeyFile:    "",
		S3SSLSkipVerify: model.Insecure.ValueBool(),
	}
	return
}
