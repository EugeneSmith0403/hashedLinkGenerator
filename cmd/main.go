package main

import (
	"context"

	"adv/go-http/configs"
	"adv/go-http/internal/account"
	"adv/go-http/internal/auth"
	"adv/go-http/internal/jwt"
	"adv/go-http/internal/link"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/payment"
	"adv/go-http/internal/payments/plan"
	"adv/go-http/internal/payments/stripe"
	"adv/go-http/internal/payments/subscription"
	"adv/go-http/internal/payments/webhook"
	stripeGo "github.com/stripe/stripe-go/v84"
	"adv/go-http/internal/stats"
	"adv/go-http/internal/user"
	"adv/go-http/pkg/db"
	"adv/go-http/pkg/event"
	"adv/go-http/pkg/middleware"
	"fmt"
	"net/http"
)

func App(config ...*configs.Config) http.Handler {
	var cfg *configs.Config
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = configs.LoadConfig()
	}
	db := db.NewDb(cfg)

	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statsRepository := stats.NewStatsRepository(db)
	eventBus := event.NewEventBus()
	statsService := stats.NewStatsService(&stats.StatServiceDep{
		EventBus:        eventBus,
		StatsRepository: statsRepository,
	})
	router := http.NewServeMux()

	//Services
	authService := auth.NewAuthService(userRepository)
	jwtService := jwt.NewJWTService(jwt.JwtDeps{
		Secret: cfg.Auth.Secret,
	})
	paymentRepository := payment.NewPaymentRepository(db)
	invoiceRepository := invoice.NewInvoiceRepository(db)
	stripeClient := stripeGo.NewClient(cfg.Stripe.ApiKey)
	stripeService := stripe.NewStripeService(stripe.StripeDeps{
		StripeClient:      stripeClient,
		WebhookSecret:     cfg.Stripe.WebhookSecret,
		ReturnURL:         cfg.Stripe.ReturnURL,
		PaymentRepository: paymentRepository,
		InvoiceRepository: invoiceRepository,
	})
	accountRepository := account.NewAccountRepository(db)
	accountService := account.NewAccountService(account.AccountServiceDeps{
		AccountRepository: accountRepository,
		PaymentService:    stripeService,
		UserRepository:    userRepository,
	})
	planRepository := plan.NewPlanRepository(db)
	subscriptionRepository := subscription.NewSubscriptionRepository(db)
	subscriptionService := subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
		SubscriptionRepository: subscriptionRepository,
		PlanRepository:         planRepository,
		StripeClient:           stripeClient,
		Ctx:                    context.Background(),
	})

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Handlers
	auth.NewAuthHandlers(router, auth.AuthHandlerDeps{
		Config:      cfg,
		AuthService: authService,
		JWTService:  jwtService,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		Config:         cfg,
		LinkRepository: linkRepository,
		JWTService:     jwtService,
		EventBus:       eventBus,
	})

	stats.NewStatsHandler(router, stats.StatsHandlerDeps{
		Config:          cfg,
		JWTService:      jwtService,
		StatsRepository: statsRepository,
	})

	account.NewAccountHandler(router, account.AccountHandlerDeps{
		AccountService: accountService,
		UserRepository: userRepository,
		JWTService:     jwtService,
	})

	stripe.NewStripeHandlers(router, stripe.StripeHandlerDeps{
		StripeService:  stripeService,
		JWTService:     jwtService,
		AccountService: accountService,
		PlanRepository: planRepository,
	})

	subscription.NewSubscriptionHandlers(router, subscription.SubscriptionHandlerDeps{
		SubscriptionService: subscriptionService,
		JWTService:          jwtService,
		AccountService:      accountService,
		PlanRepository:      planRepository,
	})

	webhook.NewWebhookHandlers(router, webhook.WebhookHandlerDeps{
		StripeService:       stripeService,
		SubscriptionService: subscriptionService,
	})

	// Middlewares
	stack := middleware.Chain(
		middleware.Cors,
		middleware.Logging,
	)

	// Events

	go statsService.AddClick()

	return stack(router)
}

func main() {

	server := http.Server{
		Addr:    ":8081",
		Handler: App(),
	}

	fmt.Println("listening 8081")
	server.ListenAndServe()
}
