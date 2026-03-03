package subscription

import (
	"time"

	stripeGo "github.com/stripe/stripe-go/v84"
)

func planPeriodEnd(name string, start time.Time) time.Time {
	switch name {
	case "Monthly":
		return start.AddDate(0, 1, 0)
	case "Semi-Annual":
		return start.AddDate(0, 6, 0)
	case "Annual":
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(0, 1, 0)
	}
}

func mapSubTimestamps(sub *stripeGo.Subscription) (cancelAt, canceledAt, trialStart, trialEnd *time.Time) {
	if sub.CancelAt != 0 {
		t := time.Unix(sub.CancelAt, 0)
		cancelAt = &t
	}
	if sub.CanceledAt != 0 {
		t := time.Unix(sub.CanceledAt, 0)
		canceledAt = &t
	}
	if sub.TrialStart != 0 {
		t := time.Unix(sub.TrialStart, 0)
		trialStart = &t
	}
	if sub.TrialEnd != 0 {
		t := time.Unix(sub.TrialEnd, 0)
		trialEnd = &t
	}
	return
}
