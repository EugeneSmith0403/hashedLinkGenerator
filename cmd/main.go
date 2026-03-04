package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"adv/go-http/configs"
	"adv/go-http/internal/account"
	"adv/go-http/internal/auth"
	"adv/go-http/internal/jwt"
	"adv/go-http/internal/link"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/payment"
	"adv/go-http/internal/payments/plan"
	"adv/go-http/internal/payments/stripe"
	stripeServices "adv/go-http/internal/payments/stripe/services"
	"adv/go-http/internal/payments/subscription"
	"adv/go-http/internal/payments/webhook"
	"adv/go-http/internal/stats"
	"adv/go-http/internal/user"
	"adv/go-http/pkg/db"
	"adv/go-http/pkg/event"
	"adv/go-http/pkg/helpers"
	"adv/go-http/pkg/middleware"
	pkgRedis "adv/go-http/pkg/redis"

	"github.com/braintree/manners"
	goRedis "github.com/go-redis/redis/v8"
	stripeGo "github.com/stripe/stripe-go/v84"
)

func loadConfigs(config ...*configs.Config) *configs.Config {
	var cfg *configs.Config
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = configs.LoadConfig()
	}

	return cfg
}

func App(cfg *configs.Config) http.Handler {
	db := db.NewDb(cfg)

	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statsRepository := stats.NewStatsRepository(db)
	eventBus := event.NewEventBus()

	// Redis
	cacheMinutes, _ := strconv.Atoi(cfg.Redis.Cache)
	redisClient := pkgRedis.NewRedis(&goRedis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	}, helpers.ToMinutes(cacheMinutes))

	statsService := stats.NewStatsService(&stats.StatServiceDep{
		EventBus:        eventBus,
		StatsRepository: statsRepository,
		RedisSrvice:     redisClient,
	})
	router := http.NewServeMux()

	// Services
	authService := auth.NewAuthService(userRepository)
	jwtService := jwt.NewJWTService(jwt.JwtDeps{
		Secret:      cfg.Auth.Secret,
		RedisSrvice: redisClient,
	})
	paymentRepository := payment.NewPaymentRepository(db)
	invoiceRepository := invoice.NewInvoiceRepository(db)
	stripeClient := stripeGo.NewClient(cfg.Stripe.ApiKey)

	customerAccountSvc := stripeServices.NewCustomerAccountService(stripeServices.CustomerAccountServiceDeps{
		StripeClient: stripeClient,
	})
	paymentSvc := stripeServices.NewPaymentService(stripeServices.PaymentServiceDeps{
		StripeClient:      stripeClient,
		WebhookSecret:     cfg.Stripe.WebhookSecret,
		ReturnURL:         cfg.Stripe.ReturnURL,
		PaymentRepository: paymentRepository,
	})
	invoiceSvc := invoice.NewInvoiceService(invoice.InvoiceServiceDeps{
		StripeClient:      stripeClient,
		InvoiceRepository: invoiceRepository,
		PaymentRepository: paymentRepository,
	})

	accountRepository := account.NewAccountRepository(db)
	accountService := account.NewAccountService(account.AccountServiceDeps{
		AccountRepository: accountRepository,
		PaymentService:    customerAccountSvc,
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
		RedisSrvice: redisClient,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		Config:         cfg,
		LinkRepository: linkRepository,
		UserRepository: userRepository,
		JWTService:     jwtService,
		EventBus:       eventBus,
	})

	stats.NewStatsHandler(router, stats.StatsHandlerDeps{
		Config:          cfg,
		JWTService:      jwtService,
		StatsRepository: statsRepository,
		StatsService:    statsService,
	})

	account.NewAccountHandler(router, account.AccountHandlerDeps{
		AccountService: accountService,
		UserRepository: userRepository,
		JWTService:     jwtService,
	})

	stripe.NewStripeHandlers(router, stripe.StripeHandlerDeps{
		PaymentService: paymentSvc,
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
		PaymentService:         paymentSvc,
		CustomerAccountService: customerAccountSvc,
		InvoiceService:         invoiceSvc,
		SubscriptionService:    subscriptionService,
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
	configs := loadConfigs()

	server := manners.NewWithServer(&http.Server{
		Addr:    ":8081",
		Handler: App(configs),
	})

	if configs.Mode == "production" {
		// Smooth shutdown
		go func() {
			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
			<-sigchan
			log.Print("Shutting down...")
			manners.Close()
		}()
	}

	log.Print("listening 8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
