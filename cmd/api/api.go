package main

import (
	"SocialMedia/internal/auth"
	"SocialMedia/internal/env"
	"SocialMedia/internal/mailer"
	"SocialMedia/internal/ratelimiter"
	"SocialMedia/internal/store"
	"SocialMedia/internal/store/cache"
	"expvar"
	"net/http"
	"time"

	"SocialMedia/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Application struct {
	config        Config
	store         store.Storage
	cacheStore    cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type MailConfig struct {
	exp       time.Duration
	fromEmail string
	sendgrid  SendGridConfig
}

type SendGridConfig struct {
	apiKey string
}

type Config struct {
	addr        string
	dbConfig    DBConfig
	mail        MailConfig
	mailer      mailer.Client
	frontendURL string
	auth        AuthConfig
	redisCfg    RedisConfig
	rateLimiter ratelimiter.Config
}

type RedisConfig struct {
	address  string
	password string
	db       int
	enable   bool
}
type AuthConfig struct {
	basic BasicAuthConfig
	token TokenConfig
}

type TokenConfig struct {
	secret   string
	audience string
	issuer   string
	exp      time.Duration
}

type BasicAuthConfig struct {
	user     string
	password string
}

type DBConfig struct {
	addr        string
	maxOpenConn int
	maxIdleConn int
	maxIdleTime string
}

func (app *Application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.RateLimiterMiddleware)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {

		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/metrics", expvar.Handler().ServeHTTP)

		// Relative to the mounted route, so the UI works behind any host or proxy.
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/v1/swagger/doc.json")))

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Post("/", app.createUserHandler)
			r.With(app.AuthTokenMiddleware(), app.RequireRoleMiddleware("admin")).Get("/", app.getUsersHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())

				r.Get("/", app.getUserByIdHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unFollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware())
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPostOwnerShip("admin", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnerShip("moderator", app.updatePostHandler))
			})
		})

		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.createUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *Application) run(mux http.Handler) error {

	// docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Title = "Social Media API"
	docs.SwaggerInfo.Description = "A social media API built with Go, Chi and PostgreSQL."

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	app.logger.Infof("Starting server at %s...", app.config.addr)

	return srv.ListenAndServe()
}
