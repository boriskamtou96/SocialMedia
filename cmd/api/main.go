package main

import (
	"SocialMedia/internal/auth"
	"SocialMedia/internal/db"
	"SocialMedia/internal/env"
	"SocialMedia/internal/mailer"
	"SocialMedia/internal/store"
	"log"
	"time"

	"go.uber.org/zap"
)

// @title        Social Media API
// @version      1.0
// @description  A social media API built with Go, Chi and PostgreSQL.
// @termsOfService http://swagger.io/terms/

// @contact.name   Boris KAMTOU
// @contact.url    https://linkedin.com/in/boriskamtou
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Paste the accessToken returned by /auth/login, prefixed with "Bearer ".
func main() {
	cfg := Config{
		addr: env.GetString("ADDR", ":8080"),
		dbConfig: DBConfig{
			addr:        env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost/socialmedia?sslmode=disable"),
			maxOpenConn: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConn: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime: env.GetString("BD_MAX_IDLE_TIME", "15m"),
		},
		mail: MailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "boriskamtou@gmail.com"),
			sendgrid: SendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		auth: AuthConfig{
			basic: BasicAuthConfig{
				user:     env.GetString("BASIC_AUTH_USER", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			token: TokenConfig{
				secret:   env.GetString("JWT_SECRET", "your-secret-key"),
				audience: env.GetString("JWT_AUDIENCE", "GopherSocialMedia"),
				issuer:   env.GetString("JWT_ISSUER", "GopherSocialMedia"),
				exp:      time.Hour * 24 * 3,
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database connection
	database, err := db.New(
		cfg.dbConfig.addr,
		cfg.dbConfig.maxOpenConn,
		cfg.dbConfig.maxIdleConn,
		cfg.dbConfig.maxIdleTime,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := database.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	logger.Info("Database connection established...")

	storage := store.NewStorage(database)

	mailer := mailer.NewSendgrid(cfg.mail.sendgrid.apiKey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.audience,
		cfg.auth.token.issuer,
	)

	app := &Application{
		config:        cfg,
		store:         storage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()
	if err := app.run(mux); err != nil {
		log.Fatal(err)
	}
}
