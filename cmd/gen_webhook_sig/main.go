package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	stripeGo "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

func main() {
	apiKey := os.Getenv("STRIPE_API_KEY")
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if apiKey == "" || secret == "" {
		fmt.Println("STRIPE_API_KEY and STRIPE_WEBHOOK_SECRET env vars are required")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <payment_intent_id> [event_type]")
		fmt.Println("Example: go run main.go pi_3xxx payment_intent.succeeded")
		os.Exit(1)
	}

	piID := os.Args[1]
	eventType := "payment_intent.succeeded"
	if len(os.Args) >= 3 {
		eventType = os.Args[2]
	}

	// Получаем реальный PaymentIntent из Stripe
	client := stripeGo.NewClient(apiKey)
	pi, err := client.V1PaymentIntents.Retrieve(context.Background(), piID, nil)
	if err != nil {
		fmt.Printf("Failed to fetch PaymentIntent: %v\n", err)
		os.Exit(1)
	}

	piJSON, err := json.Marshal(pi)
	if err != nil {
		fmt.Printf("Failed to marshal PaymentIntent: %v\n", err)
		os.Exit(1)
	}

	body := []byte(fmt.Sprintf(
		`{"id":"evt_%s","object":"event","api_version":"2026-02-25.clover","type":%q,"data":{"object":%s}}`,
		uuid.New().String(),
		eventType, piJSON,
	))

	ts := time.Now()
	signedPayload := fmt.Sprintf("%d.%s", ts.Unix(), body)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))

	sigHeader := fmt.Sprintf("t=%d,v1=%s", ts.Unix(), sig)

	_, err = webhook.ConstructEvent(body, sigHeader, secret)
	if err != nil {
		fmt.Printf("VERIFICATION FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Body:\n%s\n\n", body)
	fmt.Printf("Stripe-Signature: %s\n", sigHeader)
}
