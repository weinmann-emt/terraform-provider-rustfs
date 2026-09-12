package rustfs

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLifecycleRuleRoundTripAllActions(t *testing.T) {
	days := 60
	noncurrent := 90
	date := "2026-12-31T00:00:00Z"
	marker := true
	afterInit := 7

	rule := LifecycleRule{
		ID:     "all-actions",
		Status: "Enabled",
		Filter: LifecycleFilter{Prefix: "logs/"},
		Expiration: &LifecycleExpiration{
			Days:                      &days,
			Date:                      date,
			ExpiredObjectDeleteMarker: &marker,
		},
		Transition: &LifecycleTransition{
			Days:         &days,
			StorageClass: "WARM",
		},
		NoncurrentVersionExpiration: &LifecycleNoncurrentVersionExpiration{
			NoncurrentDays: &noncurrent,
		},
		NoncurrentVersionTransition: &LifecycleNoncurrentVersionTransition{
			NoncurrentDays: &noncurrent,
			StorageClass:   "WARM",
		},
		AbortIncompleteMultipartUpload: &LifecycleAbortIncompleteMultipartUpload{
			DaysAfterInitiation: &afterInit,
		},
	}

	config := LifecycleConfiguration{Rules: []LifecycleRule{rule}}
	body, err := xml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	wire := string(body)
	for _, want := range []string{
		"<Transition>", "<StorageClass>WARM</StorageClass>",
		"<NoncurrentVersionExpiration>", "<NoncurrentDays>90</NoncurrentDays>",
		"<NoncurrentVersionTransition>",
		"<AbortIncompleteMultipartUpload>", "<DaysAfterInitiation>7</DaysAfterInitiation>",
		"<ExpiredObjectDeleteMarker>true</ExpiredObjectDeleteMarker>",
		"<Date>2026-12-31T00:00:00Z</Date>",
	} {
		if !strings.Contains(wire, want) {
			t.Errorf("missing %q in marshalled XML:\n%s", want, wire)
		}
	}

	var got LifecycleConfiguration
	if err := xml.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := got.Rules[0]
	if r.Transition == nil || r.Transition.StorageClass != "WARM" {
		t.Error("Transition not round-tripped")
	}
	if r.NoncurrentVersionExpiration == nil || *r.NoncurrentVersionExpiration.NoncurrentDays != 90 {
		t.Error("NoncurrentVersionExpiration not round-tripped")
	}
	if r.NoncurrentVersionTransition == nil || *r.NoncurrentVersionTransition.NoncurrentDays != 90 {
		t.Error("NoncurrentVersionTransition not round-tripped")
	}
	if r.AbortIncompleteMultipartUpload == nil || *r.AbortIncompleteMultipartUpload.DaysAfterInitiation != 7 {
		t.Error("AbortIncompleteMultipartUpload not round-tripped")
	}
	if r.Expiration == nil || *r.Expiration.ExpiredObjectDeleteMarker != true {
		t.Error("ExpiredObjectDeleteMarker not round-tripped")
	}
	if r.Expiration == nil || r.Expiration.Date != date {
		t.Error("date-based Expiration not round-tripped")
	}
}

func TestLifecycleRuleRoundTripBackwardCompat(t *testing.T) {
	days := 30
	config := LifecycleConfiguration{Rules: []LifecycleRule{{
		ID:     "legacy",
		Status: "Enabled",
		Filter: LifecycleFilter{Prefix: "logs/"},
		Expiration: &LifecycleExpiration{
			Days: &days,
		},
	}}}

	body, err := xml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got LifecycleConfiguration
	if err := xml.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := got.Rules[0]
	if *r.Expiration.Days != 30 || r.Transition != nil || r.NoncurrentVersionExpiration != nil {
		t.Error("legacy prefix+days config must round-trip unchanged")
	}
}

func TestLifecycleTransitionDateRoundTrip(t *testing.T) {
	date := "2026-06-30T00:00:00Z"
	config := LifecycleConfiguration{Rules: []LifecycleRule{{
		ID:     "transition-date",
		Status: "Enabled",
		Filter: LifecycleFilter{Prefix: "archive/"},
		Transition: &LifecycleTransition{
			Date:         date,
			StorageClass: "COLD",
		},
	}}}

	body, err := xml.Marshal(config)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got LifecycleConfiguration
	if err := xml.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := got.Rules[0]
	if r.Transition == nil || r.Transition.Date != date || r.Transition.StorageClass != "COLD" {
		t.Error("transition date round-trip failed")
	}
}

func TestLifecyclePutGetRoundTrip(t *testing.T) {
	days := 60
	afterInit := 7
	var captured string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			captured = string(body)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(captured))
	}))
	defer server.Close()

	client := New(&RustfsAdminConfig{
		Endpoint:  server.Listener.Addr().String(),
		AccessKey: "admin",
	})
	client.accessSecret = "secret"

	config := &LifecycleConfiguration{Rules: []LifecycleRule{{
		ID:     "rt",
		Status: "Enabled",
		Filter: LifecycleFilter{Prefix: "p/"},
		Expiration: &LifecycleExpiration{
			Days: &days,
		},
		AbortIncompleteMultipartUpload: &LifecycleAbortIncompleteMultipartUpload{
			DaysAfterInitiation: &afterInit,
		},
	}}}

	if err := client.SetBucketLifecycleConfiguration("bucket", config); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := client.GetBucketLifecycleConfiguration("bucket")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	r := got.Rules[0]
	if r.Expiration == nil || *r.Expiration.Days != 60 {
		t.Error("expiration days not round-tripped through PUT/GET")
	}
	if r.AbortIncompleteMultipartUpload == nil || *r.AbortIncompleteMultipartUpload.DaysAfterInitiation != 7 {
		t.Error("abort incomplete multipart upload not round-tripped through PUT/GET")
	}
}
