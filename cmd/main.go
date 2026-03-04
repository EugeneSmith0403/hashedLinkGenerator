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

type repositories struct {
	link         *link.LinkRepository
	user         *user.UserRepository
	stats        *stats.StatsRepository
	payment      *payment.PaymentRepository
	invoice      *invoice.InvoiceRepository
	account      *account.AccountRepository
	plan         *plan.PlanRepository
	subscription *subscription.SubscriptionRepository
}

type services struct {
	auth         *auth.AuthService
	jwt          *jwt.JWTService
	stats        *stats.StatsService
	account      *account.AccountService
	subscription *subscription.SubscriptionService
	payment      *stripeServices.PaymentService
	customerAcct *stripeServices.CustomerAccountService
	invoice      *invoice.InvoiceService
}

type app struct {
	cfg      *configs.Config
	repos    *repositories
	svc      *services
	redis    *pkgRedis.Redis
	eventBus *event.EventBus
}

func newApp(cfg *configs.Config) *app {
	database := db.NewDb(cfg)
	eventBus := event.NewEventBus()

	cacheMinutes, _ := strconv.Atoi(cfg.Redis.Cache)
	redis := pkgRedis.NewRedis(&goRedis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	}, helpers.ToMinutes(cacheMinutes))

	repos := &repositories{
		link:         link.NewLinkRepository(database),
		user:         user.NewUserRepository(database),
		stats:        stats.NewStatsRepository(database),
		payment:      payment.NewPaymentRepository(database),
		invoice:      invoice.NewInvoiceRepository(database),
		account:      account.NewAccountRepository(database),
		plan:         plan.NewPlanRepository(database),
		subscription: subscription.NewSubscriptionRepository(database),
	}

	stripeClient := stripeGo.NewClient(cfg.Stripe.ApiKey)
	customerAcct := stripeServices.NewCustomerAccountService(stripeServices.CustomerAccountServiceDeps{
		StripeClient: stripeClient,
	})

	svc := &services{
		auth: auth.NewAuthService(repos.user),
		jwt:  jwt.NewJWTService(jwt.JwtDeps{Secret: cfg.Auth.Secret, RedisSrvice: redis}),
		stats: stats.NewStatsService(&stats.StatServiceDep{
			EventBus:        eventBus,
			StatsRepository: repos.stats,
			RedisSrvice:     redis,
		}),
		account: account.NewAccountService(account.AccountServiceDeps{
			AccountRepository: repos.account,
			PaymentService:    customerAcct,
			UserRepository:    repos.user,
		}),
		subscription: subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
			SubscriptionRepository: repos.subscription,
			PlanRepository:         repos.plan,
			StripeClient:           stripeClient,
			Ctx:                    context.Background(),
		}),
		payment: stripeServices.NewPaymentService(stripeServices.PaymentServiceDeps{
			StripeClient:      stripeClient,
			WebhookSecret:     cfg.Stripe.WebhookSecret,
			ReturnURL:         cfg.Stripe.ReturnURL,
			PaymentRepository: repos.payment,
		}),
		customerAcct: customerAcct,
		invoice: invoice.NewInvoiceService(invoice.InvoiceServiceDeps{
			StripeClient:      stripeClient,
			InvoiceRepository: repos.invoice,
			PaymentRepository: repos.payment,
		}),
	}

	return &app{cfg: cfg, repos: repos, svc: svc, redis: redis, eventBus: eventBus}
}

func (a *app) registerHandlers(router *http.ServeMux) {
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	auth.NewAuthHandlers(router, auth.AuthHandlerDeps{
		Config:      a.cfg,
		AuthService: a.svc.auth,
		JWTService:  a.svc.jwt,
		RedisSrvice: a.redis,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		Config:         a.cfg,
		LinkRepository: a.repos.link,
		UserRepository: a.repos.user,
		JWTService:     a.svc.jwt,
		EventBus:       a.eventBus,
	})
	stats.NewStatsHandler(router, stats.StatsHandlerDeps{
		Config:          a.cfg,
		JWTService:      a.svc.jwt,
		StatsRepository: a.repos.stats,
		StatsService:    a.svc.stats,
	})
	account.NewAccountHandler(router, account.AccountHandlerDeps{
		AccountService: a.svc.account,
		UserRepository: a.repos.user,
		JWTService:     a.svc.jwt,
	})
	stripe.NewStripeHandlers(router, stripe.StripeHandlerDeps{
		PaymentService: a.svc.payment,
		JWTService:     a.svc.jwt,
		AccountService: a.svc.account,
		PlanRepository: a.repos.plan,
	})
	subscription.NewSubscriptionHandlers(router, subscription.SubscriptionHandlerDeps{
		SubscriptionService: a.svc.subscription,
		JWTService:          a.svc.jwt,
		AccountService:      a.svc.account,
		PlanRepository:      a.repos.plan,
	})
	webhook.NewWebhookHandlers(router, webhook.WebhookHandlerDeps{
		PaymentService:         a.svc.payment,
		CustomerAccountService: a.svc.customerAcct,
		InvoiceService:         a.svc.invoice,
		SubscriptionService:    a.svc.subscription,
	})
}

func App(cfg *configs.Config) http.Handler {
	a := newApp(cfg)
	router := http.NewServeMux()
	a.registerHandlers(router)
	go a.svc.stats.AddClick()
	return middleware.Chain(middleware.Cors, middleware.Logging)(router)
}

func loadConfigs(config ...*configs.Config) *configs.Config {
	if len(config) > 0 {
		return config[0]
	}
	return configs.LoadConfig()
}

func main() {
	configs := loadConfigs()

	server := manners.NewWithServer(&http.Server{
		Addr:    ":8081",
		Handler: App(configs),
	})

	if configs.Mode == "production" {
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
