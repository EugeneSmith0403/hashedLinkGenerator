package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"link-generator/api"
	"link-generator/cmd/shared"
	"link-generator/configs"
	"link-generator/internal/account"
	"link-generator/internal/auth"
	authsession "link-generator/internal/auth_session"
	"link-generator/internal/jwt"
	"link-generator/internal/link"
	"link-generator/internal/locales"
	"link-generator/internal/mailer"
	"link-generator/internal/models"
	"link-generator/internal/payments/invoice"
	"link-generator/internal/payments/payment"
	"link-generator/internal/payments/plan"
	"link-generator/internal/payments/stripe"
	stripeServices "link-generator/internal/payments/stripe/services"
	"link-generator/internal/payments/subscription"
	"link-generator/internal/payments/webhook"
	"link-generator/internal/stats"
	"link-generator/internal/user"
	pkgClickhouse "link-generator/pkg/clickhouse"
	"github.com/oschwald/geoip2-golang"
	"link-generator/pkg/db"
	"link-generator/pkg/event"
	"link-generator/pkg/helpers"
	"link-generator/pkg/limiter"
	"link-generator/pkg/middleware"
	rabbitmq "link-generator/pkg/rabbitMq"
	pkgRedis "link-generator/pkg/redis"

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
	authSession  *authsession.AuthSessionService
}

type app struct {
	cfg           *configs.Config
	db            *db.Db
	repos         *repositories
	svc           *services
	redis         *pkgRedis.Redis
	eventBus      *event.EventBus
	rabbitMq      *rabbitmq.RabbitMq
	clickhouse    *pkgClickhouse.Clickhouse
	geoIP         *geoip2.Reader
	rateLimiter   *limiter.LimiterService
	ipRateLimiter *limiter.LimiterService
}

type subscriptionUserAdapter struct {
	svc *subscription.SubscriptionService
}

func newTestHandler(cfg *configs.Config) http.Handler {
	a := newApp(cfg)
	router := http.NewServeMux()
	a.registerHandlers(router)
	return middleware.Chain(middleware.Cors, middleware.Logging)(router)
}

func newApp(cfg *configs.Config) *app {
	database := db.NewDb(cfg)
	eventBus := event.NewEventBus()
	rabbitMq := rabbitmq.NewRabbitMq(cfg.RabbitMq)

	ch, err := pkgClickhouse.NewCliсkhouse(&cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse init: %v", err)
	}

	var geoIPReader *geoip2.Reader
	if cfg.GeoIPPath != "" {
		geoIPReader, err = geoip2.Open(cfg.GeoIPPath)
		if err != nil {
			log.Printf("geoip init: %v (country lookup disabled)", err)
		}
	}

	cacheMinutes, _ := strconv.Atoi(cfg.Redis.Cache)
	redis := pkgRedis.NewRedis(&goRedis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	}, helpers.ToMinutes(cacheMinutes))

	repos := &repositories{
		link:         link.NewLinkRepository(database),
		user:         user.NewUserRepository(database),
		stats:        stats.NewStatsRepository(ch),
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

	jwtService := jwt.NewJWTService(jwt.JwtDeps{Secret: cfg.Auth.Secret, RedisSrvice: redis})

	svc := &services{
		auth: auth.NewAuthService(auth.AuthServiceDeps{
			UserRepository: repos.user,
			Config:         cfg,
			JWTService:     jwtService,
			RedisService:   redis,
		}),
		jwt: jwtService,
		stats: stats.NewStatsService(&stats.StatServiceDep{
			EventBus:        eventBus,
			StatsRepository: repos.stats,
			GeoIP:           geoIPReader,
		}),
		account: account.NewAccountService(account.AccountServiceDeps{
			AccountRepository: repos.account,
			PaymentService:    customerAcct,
			UserRepository:    repos.user,
			Redis:             redis,
		}),
		subscription: subscription.NewSubscriptionService(subscription.SubscriptionServiceDeps{
			SubscriptionRepository: repos.subscription,
			PlanRepository:         repos.plan,
			PaymentRepository:      repos.payment,
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
			StripeClient:           stripeClient,
			InvoiceRepository:      repos.invoice,
			PaymentRepository:      repos.payment,
			SubscriptionRepository: repos.subscription,
		}),
		authSession: authsession.NewAuthSessionService(database),
	}

	rdb := redis.Client()
	accountLimiter := limiter.NewLimiter(rdb, limiter.Config{Capacity: 100, RefillRate: 100, KeyType: limiter.KeyByAccountID})
	ipLimiter := limiter.NewLimiter(rdb, limiter.Config{Capacity: 100, RefillRate: 100, KeyType: limiter.KeyByIP})

	return &app{cfg: cfg, db: database, repos: repos, svc: svc, redis: redis, eventBus: eventBus, rabbitMq: rabbitMq, clickhouse: ch, geoIP: geoIPReader, rateLimiter: accountLimiter, ipRateLimiter: ipLimiter}
}

func (a *subscriptionUserAdapter) GetSubscriptionByUserID(userID uint) (*models.SubscriptionInfo, error) {
	sub, err := a.svc.GetSubscriptionByUserId(userID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}
	return &models.SubscriptionInfo{
		ID:                 sub.ID,
		CreatedAt:          sub.CreatedAt,
		PlanID:             sub.PlanID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAt:           sub.CancelAt,
		CanceledAt:         sub.CanceledAt,
		TrialStart:         sub.TrialStart,
		TrialEnd:           sub.TrialEnd,
		IsPaymentIntent:    strings.HasPrefix(sub.BillingID, "pi_"),
	}, nil
}

func (a *app) registerHandlers(router *http.ServeMux) {
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	auth.NewAuthHandlers(router, auth.AuthHandlerDeps{
		AuthService: a.svc.auth,
		AuthMailerDeps: auth.AuthMailerDeps{
			Mailer: mailer.NewMailer(mailer.MailerDeps{
				LocalesFS:  locales.FS,
				LocalesDir: "auth/register",
				Host:       a.cfg.Mailer.Host,
				Port:       a.cfg.Mailer.Port,
				User:       a.cfg.Mailer.User,
				Password:   a.cfg.Mailer.Password,
			}),
			MailerFrom: a.cfg.Mailer.From,
			AppName:    "Go Adv",
			AppURL:     a.cfg.Stripe.ReturnURL,
		},
		AccountService:     a.svc.account,
		AuthSeseionService: a.svc.authSession,
		IPRateLimiter:      a.ipRateLimiter,
		UserRepository:     a.repos.user,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		Config:              a.cfg,
		LinkRepository:      a.repos.link,
		UserRepository:      a.repos.user,
		AuthSessionService:  a.svc.authSession,
		EventBus:            a.eventBus,
		SubscriptionService: a.svc.subscription,
		RabbitMq:            a.rabbitMq,
		StatsService:        a.svc.stats,
		RateLimiter:         a.rateLimiter,
		IPRateLimiter:       a.ipRateLimiter,
	})
	stats.NewStatsHandler(router, stats.StatsHandlerDeps{
		Config:             a.cfg,
		AuthSessionService: a.svc.authSession,
		StatsRepository:    a.repos.stats,
		StatsService:       a.svc.stats,
		Redis:              a.redis,
		RateLimiter:        a.rateLimiter,
	})
	account.NewAccountHandler(router, account.AccountHandlerDeps{
		AccountService:     a.svc.account,
		UserRepository:     a.repos.user,
		AuthSessionService: a.svc.authSession,
		RateLimiter:        a.rateLimiter,
	})
	payment.NewPaymentHandler(router, payment.PaymentHandlerDeps{
		PaymentRepository:  a.repos.payment,
		AuthSessionService: a.svc.authSession,
		AccountService:     a.svc.account,
		RateLimiter:        a.rateLimiter,
	})
	stripe.NewStripeHandlers(router, stripe.StripeHandlerDeps{
		PaymentService:      a.svc.payment,
		AuthSessionService:  a.svc.authSession,
		AccountService:      a.svc.account,
		PlanRepository:      a.repos.plan,
		SubscriptionService: a.svc.subscription,
		RateLimiter:         a.rateLimiter,
	})
	subscription.NewSubscriptionHandlers(router, subscription.SubscriptionHandlerDeps{
		SubscriptionService: a.svc.subscription,
		AuthSessionService:  a.svc.authSession,
		AccountService:      a.svc.account,
		PlanRepository:      a.repos.plan,
		RateLimiter:         a.rateLimiter,
	})
	plan.NewPlanHandler(router, plan.PlanHandlerDeps{
		PlanRepository:     a.repos.plan,
		Redis:              a.redis,
		AuthSessionService: a.svc.authSession,
		RateLimiter:        a.rateLimiter,
	})
	webhook.NewWebhookHandlers(router, webhook.WebhookHandlerDeps{
		PaymentService:         a.svc.payment,
		CustomerAccountService: a.svc.customerAcct,
		InvoiceService:         a.svc.invoice,
		SubscriptionService:    a.svc.subscription,
		AccountRepository:      a.repos.account,
		RabbitMq:               a.rabbitMq,
	})
	user.NewUserHandlers(router, user.UserHandlerDeps{
		UserRepository:      a.repos.user,
		AccountService:      a.svc.account,
		SubscriptionService: &subscriptionUserAdapter{svc: a.svc.subscription},
		AuthService:         a.svc.auth,
		AuthSeseionService:  a.svc.authSession,
		RateLimiter:         a.rateLimiter,
		IPRateLimiter:       a.ipRateLimiter,
	})

	api.RegisterDocsRoutes(router, "api/openapi.yaml")
}

func main() {
	configs := shared.LoadConfigs()

	a := newApp(configs)
	defer a.redis.Close()
	defer a.rabbitMq.Close()
	defer a.db.Close()
	defer a.clickhouse.Close()
	if a.geoIP != nil {
		defer a.geoIP.Close()
	}

	router := http.NewServeMux()
	a.registerHandlers(router)
	handler := middleware.Chain(middleware.Cors, middleware.Logging)(router)

	server := manners.NewWithServer(&http.Server{
		Addr:    ":8081",
		Handler: handler,
	})

	if configs.Mode == "production" {
		go func() {
			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
			<-sigchan
			log.Print("Shutting down...")
			server.Close()
		}()
	}

	log.Print("listening 8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
