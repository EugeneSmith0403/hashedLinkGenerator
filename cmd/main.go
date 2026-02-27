package main

import (
	"adv/go-http/configs"
	"adv/go-http/internal/auth"
	"adv/go-http/internal/jwt"
	"adv/go-http/internal/link"
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
