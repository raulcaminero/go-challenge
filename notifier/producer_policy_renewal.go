package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gloveboxhq/glovebox-go-code-challenge/notifier/channels/email"
)

const TopicPolicyRenewal = "policy-renewal"

// renewalDateLayout is the date format handed to the email template. Templates
// receive a plain string rather than a time.Time so the rendering layer never
// depends on Go-specific types.
const renewalDateLayout = "2006-01-02"

// PolicyRenewalInput is the business input for a renewal reminder.
type PolicyRenewalInput struct {
	Recipient    string
	PolicyNumber string
	RenewalDate  time.Time
}

// policyRenewalTopicBuilder turns a PolicyRenewalInput into an email Request.
//
// now is injectable so the "renewal already passed" rule is deterministic in
// tests; the zero value falls back to time.Now.
type policyRenewalTopicBuilder struct {
	now func() time.Time
}

func (policyRenewalTopicBuilder) Topic() string { return TopicPolicyRenewal }

func (b policyRenewalTopicBuilder) BuildRequest(_ context.Context, input any) (Request, error) {
	typedInput, ok := input.(PolicyRenewalInput)
	if !ok {
		return Request{}, fmt.Errorf("invalid policy renewal input type")
	}
	if typedInput.Recipient == "" {
		return Request{}, fmt.Errorf("policy renewal requires recipient")
	}
	if typedInput.PolicyNumber == "" {
		return Request{}, fmt.Errorf("policy renewal requires policy number")
	}
	if typedInput.RenewalDate.IsZero() {
		return Request{}, fmt.Errorf("policy renewal requires renewal date")
	}

	// A reminder for a renewal that has already passed is a data problem
	// upstream, not something we should email a policyholder about. Compare at
	// day granularity so a renewal due today is still sent.
	//
	// Both dates are evaluated in the renewal's own time zone: the calendar
	// day the policyholder sees on their documents is the one that matters,
	// not the UTC day the notification happens to be produced on.
	loc := typedInput.RenewalDate.Location()
	today := truncateToDay(b.currentTime().In(loc))
	renewal := truncateToDay(typedInput.RenewalDate)
	if renewal.Before(today) {
		return Request{}, fmt.Errorf("policy renewal date %s is in the past", renewal.Format(renewalDateLayout))
	}
	daysUntilRenewal := daysBetween(today, renewal)

	return Request{
		Topic:      TopicPolicyRenewal,
		Recipients: []string{typedInput.Recipient},
		Template:   email.TplPolicyRenewal,
		Vars: map[string]any{
			"policyNumber":     typedInput.PolicyNumber,
			"renewalDate":      renewal.Format(renewalDateLayout),
			"daysUntilRenewal": daysUntilRenewal,
		},
	}, nil
}

func (b policyRenewalTopicBuilder) currentTime() time.Time {
	if b.now != nil {
		return b.now()
	}
	return time.Now()
}

// truncateToDay drops the time-of-day component, keeping t's location, so
// date comparisons are not affected by the hour the notification is produced.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// daysBetween counts calendar days from a to b (both already truncated to a
// day in the same location). Rounding guards against DST transitions, where a
// "day" is 23 or 25 hours long.
func daysBetween(a, b time.Time) int {
	return int(math.Round(b.Sub(a).Hours() / 24))
}
