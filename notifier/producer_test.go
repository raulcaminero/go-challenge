package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gloveboxhq/glovebox-go-code-challenge/notifier/channels/email"
	"github.com/gloveboxhq/glovebox-go-code-challenge/notifier/channels/email/mockemail"
)

func TestNotifyDocumentUpload(t *testing.T) {
	mail := mockemail.NewClient()
	producer := NewProducer(mail)

	err := producer.NotifyTopic(context.Background(), TopicDocumentUpload, DocumentUploadInput{
		Recipient: "user@example.com",
		Document:  "policy.pdf",
	})
	if err != nil {
		t.Fatalf("notify document upload: %v", err)
	}

	logs := mail.SendLogs()
	if logs.IsEmpty() {
		t.Fatal("expected send log")
	}
	last := logs.Last()
	if last.To != "user@example.com" {
		t.Fatalf("expected recipient %q, got %q", "user@example.com", last.To)
	}
	if last.Tpl != email.TplDocumentUpload {
		t.Fatalf("expected template %q, got %q", email.TplDocumentUpload, last.Tpl)
	}
}

func TestNotifyUnknownTopic(t *testing.T) {
	err := NewProducer(mockemail.NewClient()).NotifyTopic(context.Background(), "unknown", nil)
	if err == nil {
		t.Fatal("expected unknown topic error")
	}
}

// fixedNow pins "today" so the past-date rule is deterministic.
var fixedNow = time.Date(2026, time.March, 10, 15, 30, 0, 0, time.UTC)

// tokyo is a fixed +09:00 zone (no tzdata dependency) used to prove dates are
// evaluated in the renewal's own location.
var tokyo = time.FixedZone("Asia/Tokyo", 9*60*60)

func TestPolicyRenewalBuildRequest(t *testing.T) {
	builder := policyRenewalTopicBuilder{now: func() time.Time { return fixedNow }}

	valid := PolicyRenewalInput{
		Recipient:    "holder@example.com",
		PolicyNumber: "POL-2026-0042",
		RenewalDate:  fixedNow.AddDate(0, 0, 30),
	}

	tests := map[string]struct {
		input    any
		wantErr  bool
		wantVars map[string]any
	}{
		"success": {
			input: valid,
			wantVars: map[string]any{
				"policyNumber":     "POL-2026-0042",
				"renewalDate":      "2026-04-09",
				"daysUntilRenewal": 30,
			},
		},
		"renewal due today is still sent": {
			// Different hour than fixedNow on the same day: comparison is per day.
			input: PolicyRenewalInput{Recipient: valid.Recipient, PolicyNumber: valid.PolicyNumber,
				RenewalDate: time.Date(2026, time.March, 10, 1, 0, 0, 0, time.UTC)},
			wantVars: map[string]any{
				"policyNumber":     "POL-2026-0042",
				"renewalDate":      "2026-03-10",
				"daysUntilRenewal": 0,
			},
		},
		"renewal date keeps its own calendar day east of UTC": {
			// Local midnight in Tokyo is the previous day in UTC. The email must
			// show the day printed on the policyholder's documents (2026-03-20)
			// and count days in that zone, not shift it to 2026-03-19.
			input: PolicyRenewalInput{Recipient: valid.Recipient, PolicyNumber: valid.PolicyNumber,
				RenewalDate: time.Date(2026, time.March, 20, 0, 0, 0, 0, tokyo)},
			wantVars: map[string]any{
				"policyNumber":     "POL-2026-0042",
				"renewalDate":      "2026-03-20",
				"daysUntilRenewal": 9, // fixedNow (Mar 10 15:30 UTC) is Mar 11 00:30 in Tokyo
			},
		},
		"wrong input type": {
			input:   DocumentUploadInput{Recipient: "x@example.com", Document: "a.pdf"},
			wantErr: true,
		},
		"missing recipient": {
			input:   PolicyRenewalInput{PolicyNumber: valid.PolicyNumber, RenewalDate: valid.RenewalDate},
			wantErr: true,
		},
		"missing policy number": {
			input:   PolicyRenewalInput{Recipient: valid.Recipient, RenewalDate: valid.RenewalDate},
			wantErr: true,
		},
		"missing renewal date": {
			input:   PolicyRenewalInput{Recipient: valid.Recipient, PolicyNumber: valid.PolicyNumber},
			wantErr: true,
		},
		"renewal date in the past": {
			input: PolicyRenewalInput{Recipient: valid.Recipient, PolicyNumber: valid.PolicyNumber,
				RenewalDate: fixedNow.AddDate(0, 0, -1)},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := builder.BuildRequest(context.Background(), tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got request %+v", req)
				}
				return
			}
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			if req.Topic != TopicPolicyRenewal {
				t.Fatalf("topic = %q, want %q", req.Topic, TopicPolicyRenewal)
			}
			if req.Template != email.TplPolicyRenewal {
				t.Fatalf("template = %q, want %q", req.Template, email.TplPolicyRenewal)
			}
			if len(req.Recipients) != 1 || req.Recipients[0] != valid.Recipient {
				t.Fatalf("recipients = %v, want [%s]", req.Recipients, valid.Recipient)
			}
			if !reflect.DeepEqual(req.Vars, tc.wantVars) {
				t.Fatalf("vars = %#v, want %#v", req.Vars, tc.wantVars)
			}
		})
	}
}

// TestNotifyPolicyRenewal exercises the registered topic end to end through
// the producer and the mock mail channel.
func TestNotifyPolicyRenewal(t *testing.T) {
	mail := mockemail.NewClient()
	producer := NewProducer(mail)

	err := producer.NotifyTopic(context.Background(), TopicPolicyRenewal, PolicyRenewalInput{
		Recipient:    "holder@example.com",
		PolicyNumber: "POL-2026-0042",
		RenewalDate:  time.Now().AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatalf("notify policy renewal: %v", err)
	}

	logs := mail.SendLogs()
	if len(logs) != 1 {
		t.Fatalf("expected exactly one send log, got %d", len(logs))
	}
	last := logs.Last()
	if last.To != "holder@example.com" {
		t.Fatalf("expected recipient %q, got %q", "holder@example.com", last.To)
	}
	if last.Tpl != email.TplPolicyRenewal {
		t.Fatalf("expected template %q, got %q", email.TplPolicyRenewal, last.Tpl)
	}
	if last.Vars["policyNumber"] != "POL-2026-0042" {
		t.Fatalf("expected policyNumber var, got %v", last.Vars)
	}
}

// TestNotifyPolicyRenewalInvalidInputSendsNothing proves validation failures
// never reach the mail channel.
func TestNotifyPolicyRenewalInvalidInputSendsNothing(t *testing.T) {
	mail := mockemail.NewClient()
	err := NewProducer(mail).NotifyTopic(context.Background(), TopicPolicyRenewal, PolicyRenewalInput{
		PolicyNumber: "POL-1",
		RenewalDate:  time.Now().AddDate(0, 1, 0),
	})
	if err == nil {
		t.Fatal("expected validation error for missing recipient")
	}
	if !mail.SendLogs().IsEmpty() {
		t.Fatalf("expected no send logs, got %d", len(mail.SendLogs()))
	}
}

// TestRegisteredTopicsMatchKeys guards the registry: every builder must be
// reachable under its own Topic().
func TestRegisteredTopicsMatchKeys(t *testing.T) {
	producer := NewProducer(mockemail.NewClient())
	for key, builder := range producer.topicBuilders {
		if builder.Topic() != key {
			t.Errorf("builder registered under %q reports topic %q", key, builder.Topic())
		}
	}
	for _, topic := range []string{TopicDocumentUpload, TopicOTPLogin, TopicPolicyRenewal} {
		if _, ok := producer.topicBuilders[topic]; !ok {
			t.Errorf("topic %q is not registered", topic)
		}
	}
}
