package main

import (
	"SocialMedia/internal/db"
	"SocialMedia/internal/env"
	"SocialMedia/internal/store"
	"log"
)

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
