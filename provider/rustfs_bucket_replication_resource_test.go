package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildReplicationConfig_basic(t *testing.T) {
	plan := bucketReplicationResourceModel{
		Bucket:            types.StringValue("source-bucket"),
		Role:              types.StringValue("arn:minio:replication::id:src"),
		DestinationBucket: types.StringValue("arn:aws:s3:::dest"),
		Priority:          types.Int64Value(1),
		Status:            types.StringValue("Enabled"),
	}

	cfg := buildReplicationConfig(plan)

	if cfg.Role != "arn:minio:replication::id:src" {
		t.Errorf("expected role, got %s", cfg.Role)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Status != "Enabled" {
		t.Errorf("expected Enabled, got %s", cfg.Rules[0].Status)
	}
	if cfg.Rules[0].Destination.Bucket != "arn:aws:s3:::dest" {
		t.Errorf("unexpected dest: %s", cfg.Rules[0].Destination.Bucket)
	}
}

func TestBuildReplicationConfig_deleteReplication(t *testing.T) {
	plan := bucketReplicationResourceModel{
		Bucket:                 types.StringValue("source"),
		Role:                   types.StringValue("arn:minio:replication::id:src"),
		DestinationBucket:      types.StringValue("arn:aws:s3:::dest"),
		Priority:               types.Int64Value(1),
		Status:                 types.StringValue("Enabled"),
		DeleteMarkerReplication: types.StringValue("Enabled"),
		DeleteReplication:      types.StringValue("Disabled"),
	}

	cfg := buildReplicationConfig(plan)
	rule := cfg.Rules[0]

	if rule.DeleteMarkerReplication.Status != "Enabled" {
		t.Errorf("expected Enabled, got %s", rule.DeleteMarkerReplication.Status)
	}
	if rule.DeleteReplication.Status != "Disabled" {
		t.Errorf("expected Disabled, got %s", rule.DeleteReplication.Status)
	}
}
