package main

import (
	"SocialMedia/internal/db"
	"SocialMedia/internal/env"
	"SocialMedia/internal/store"
	"log"
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
	}

	database, err := db.New(
		cfg.dbConfig.addr,
		cfg.dbConfig.maxOpenConn,
		cfg.dbConfig.maxIdleConn,
		cfg.dbConfig.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		err := database.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	log.Println("Database connection established...")

	storage := store.NewStorage(database)

	app := &Application{
		config: cfg,
		store:  storage,
	}

	mux := app.mount()
	if err := app.run(mux); err != nil {
		log.Fatal(err)
	}
}
