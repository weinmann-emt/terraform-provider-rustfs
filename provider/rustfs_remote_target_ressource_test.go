package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// remoteTargetPeer returns the peer connection settings required to exercise
// rustfs_remote_target. Because set-remote-target validates that the remote
// endpoint is reachable and the destination bucket exists, a second running
// rustfs peer is required. When it is not configured the test is skipped.
type remoteTargetPeer struct {
	endpoint   string // endpoint reachable from the rustfs server
	s3Endpoint string // endpoint reachable from the test host for bucket setup
	accessKey  string
	secretKey  string
}

func remoteTargetPeerConfig(t *testing.T) (*remoteTargetPeer, error) {
	t.Helper()
	p := &remoteTargetPeer{
		endpoint:  os.Getenv("RUSTFS_PEER_ENDPOINT"),
		accessKey: os.Getenv("RUSTFS_PEER_ACCESS_KEY"),
		secretKey: os.Getenv("RUSTFS_PEER_SECRET_KEY"),
	}
	if p.endpoint == "" || p.accessKey == "" || p.secretKey == "" {
		return nil, fmt.Errorf("RUSTFS_PEER_ENDPOINT, RUSTFS_PEER_ACCESS_KEY and RUSTFS_PEER_SECRET_KEY must be set")
	}
	if v := os.Getenv("RUSTFS_PEER_S3_ENDPOINT"); v != "" {
		p.s3Endpoint = v
	} else {
		p.s3Endpoint = p.endpoint
	}
	return p, nil
}

func TestRemoteTargetRessourceSchema(t *testing.T) {
	r := NewRemoteTargetRessource()
	resp := &fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"arn", "type", "endpoint", "access_key", "secret_key", "secure", "region", "path", "bucket", "target_bucket"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing schema attribute %q", attr)
		}
	}
}

func TestAccRemoteTargetResource(t *testing.T) {
	peer, err := remoteTargetPeerConfig(t)
	if err != nil {
		t.Skipf("skipping remote target acceptance test: %v", err)
	}

	// The source bucket must be versioned and the destination bucket must
	// exist on the (versioned) peer for set-remote-target to succeed.
	srcBucket := fmt.Sprintf("tf-rt-src-%d", acctest.RandInt())
	dstBucket := fmt.Sprintf("tf-rt-dst-%d", acctest.RandInt())

	srcClient, err := testAccMinioClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := srcClient.MakeBucket(context.Background(), srcBucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatalf("creating source bucket: %v", err)
	}
	t.Cleanup(func() { _ = srcClient.RemoveBucket(context.Background(), srcBucket) })
	if err := srcClient.SetBucketVersioning(context.Background(), srcBucket, minio.BucketVersioningConfiguration{Status: "Enabled"}); err != nil {
		t.Fatalf("enabling versioning on source bucket: %v", err)
	}

	peerClient, err := minio.New(peer.s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(peer.accessKey, peer.secretKey, ""),
		Secure: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := peerClient.MakeBucket(context.Background(), dstBucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatalf("creating destination bucket on peer: %v", err)
	}
	t.Cleanup(func() { _ = peerClient.RemoveBucket(context.Background(), dstBucket) })
	if err := peerClient.SetBucketVersioning(context.Background(), dstBucket, minio.BucketVersioningConfiguration{Status: "Enabled"}); err != nil {
		t.Fatalf("enabling versioning on destination bucket: %v", err)
	}

	resourceName := "rustfs_remote_target.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRemoteTargetConfig(srcBucket, dstBucket, peer, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "replication"),
					resource.TestCheckResourceAttr(resourceName, "bucket", srcBucket),
					resource.TestCheckResourceAttr(resourceName, "target_bucket", dstBucket),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					return rs.Primary.Attributes["bucket"] + ":" + rs.Primary.Attributes["arn"], nil
				},
				ImportStateVerifyIdentifierAttribute: "arn",
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"secret_key"},
			},
			{
				Config: testAccRemoteTargetConfig(srcBucket, dstBucket, peer, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "secure", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func testAccRemoteTargetConfig(srcBucket, dstBucket string, peer *remoteTargetPeer, secure string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "rustfs_remote_target" "test" {
  type          = "replication"
  endpoint      = "%s"
  access_key    = "%s"
  secret_key    = "%s"
  secure        = %s
  bucket        = "%s"
  target_bucket = "%s"
}
`, peer.endpoint, peer.accessKey, peer.secretKey, secure, srcBucket, dstBucket)
}
