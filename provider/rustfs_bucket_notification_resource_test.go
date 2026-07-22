package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildNotificationConfig_SingleQueue(t *testing.T) {
	eventsSet, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"})
	plan := bucketNotificationResourceModel{
		Bucket: types.StringValue("test-bucket"),
		Queue: []bucketNotificationQueueModel{
			{
				Arn:          types.StringValue("arn:minio:sqs::PRIMARY:amqp"),
				Events:       eventsSet,
				FilterPrefix: types.StringValue("uploads/"),
				FilterSuffix: types.StringValue(".jpg"),
			},
		},
	}

	config := buildNotificationConfig(plan)

	if len(config.QueueConfigs) != 1 {
		t.Fatalf("expected 1 queue config, got %d", len(config.QueueConfigs))
	}

	q := config.QueueConfigs[0]
	if q.Queue != "arn:minio:sqs::PRIMARY:amqp" {
		t.Errorf("unexpected queue ARN: %s", q.Queue)
	}
	if len(q.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(q.Events))
	}

	if q.Filter == nil {
		t.Fatal("expected non-nil filter")
	}

	rules := q.Filter.S3Key.FilterRules
	if len(rules) != 2 {
		t.Fatalf("expected 2 filter rules, got %d", len(rules))
	}

	foundPrefix, foundSuffix := false, false
	for _, r := range rules {
		if r.Name == "prefix" && r.Value == "uploads/" {
			foundPrefix = true
		}
		if r.Name == "suffix" && r.Value == ".jpg" {
			foundSuffix = true
		}
	}
	if !foundPrefix {
		t.Error("missing prefix filter rule")
	}
	if !foundSuffix {
		t.Error("missing suffix filter rule")
	}
}

func TestBuildNotificationConfig_NoFilter(t *testing.T) {
	eventsSet, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"s3:ObjectCreated:*"})
	plan := bucketNotificationResourceModel{
		Bucket: types.StringValue("test-bucket"),
		Queue: []bucketNotificationQueueModel{
			{
				Arn:    types.StringValue("arn:minio:sqs::PRIMARY:amqp"),
				Events: eventsSet,
			},
		},
	}

	config := buildNotificationConfig(plan)
	if len(config.QueueConfigs) != 1 {
		t.Fatalf("expected 1 queue config, got %d", len(config.QueueConfigs))
	}
	if config.QueueConfigs[0].Filter != nil {
		t.Error("expected nil filter when no prefix/suffix")
	}
}

func TestBuildNotificationConfig_MultipleQueues(t *testing.T) {
	events1, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"s3:ObjectCreated:*"})
	events2, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"s3:ObjectRemoved:*"})
	plan := bucketNotificationResourceModel{
		Bucket: types.StringValue("test-bucket"),
		Queue: []bucketNotificationQueueModel{
			{Arn: types.StringValue("arn:minio:sqs::PRIMARY:q1"), Events: events1},
			{Arn: types.StringValue("arn:minio:sqs::PRIMARY:q2"), Events: events2},
		},
	}

	config := buildNotificationConfig(plan)
	if len(config.QueueConfigs) != 2 {
		t.Fatalf("expected 2 queue configs, got %d", len(config.QueueConfigs))
	}
}
