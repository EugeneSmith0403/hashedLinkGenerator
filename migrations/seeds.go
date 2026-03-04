package main

import (
	"context"
	"fmt"

	"adv/go-http/internal/payments/plan"

	stripe "github.com/stripe/stripe-go/v84"
	stripeGo "github.com/stripe/stripe-go/v84"
	"gorm.io/gorm"
)

type stripeInterval struct {
	Interval      string
	IntervalCount int64
}

var planIntervals = map[string]stripeInterval{
	"Monthly":     {Interval: "month", IntervalCount: 1},
	"Semi-Annual": {Interval: "month", IntervalCount: 6},
	"Annual":      {Interval: "year", IntervalCount: 1},
}

func seedPlans(db *gorm.DB, stripeClient *stripeGo.Client) {
	ctx := context.Background()

	plans := []plan.Plan{
		{
			Name:     "Monthly",
			Cost:     20,
			Currency: "usd",
			IsActive: true,
		},
		{
			Name:     "Semi-Annual",
			Cost:     40,
			Currency: "usd",
			IsActive: true,
		},
		{
			Name:     "Annual",
			Cost:     100,
			Currency: "usd",
			IsActive: true,
		},
	}

	for _, p := range plans {
		var existing plan.Plan
		result := db.Where(plan.Plan{Name: p.Name}).FirstOrCreate(&existing, p)
		if result.Error != nil {
			fmt.Printf("failed to find or create plan %s: %v\n", p.Name, result.Error)
			continue
		}

		if existing.StripePriceID != "" {
			continue
		}

		product, err := stripeClient.V1Products.Create(ctx, &stripeGo.ProductCreateParams{
			Name: stripe.String(p.Name),
		})
		if err != nil {
			fmt.Printf("failed to create stripe product for plan %s: %v\n", p.Name, err)
			continue
		}

		price, err := stripeClient.V1Prices.Create(ctx, &stripeGo.PriceCreateParams{
			Product:    stripe.String(product.ID),
			UnitAmount: stripe.Int64(int64(p.Cost * 100)),
			Currency:   stripe.String(p.Currency),
			Recurring: &stripeGo.PriceCreateRecurringParams{
				Interval:      stripe.String(planIntervals[p.Name].Interval),
				IntervalCount: stripe.Int64(planIntervals[p.Name].IntervalCount),
			},
		})
		if err != nil {
			fmt.Printf("failed to create stripe price for plan %s: %v\n", p.Name, err)
			continue
		}

		db.Model(&existing).Updates(plan.Plan{
			StripeProductID: product.ID,
			StripePriceID:   price.ID,
		})

		fmt.Printf("plan %s: product=%s price=%s\n", p.Name, product.ID, price.ID)
	}
}
