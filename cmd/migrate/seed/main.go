package main

import (
	"SocialMedia/internal/db"
	store2 "SocialMedia/internal/store"
)

func main() {
	conn, err := db.New("postgres://postgres:postgres@localhost:5432/socialmedia?sslmode=disable", 10, 5, "5m")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Seed the database with initial data here
	// For example, you can create users, posts, and comments
	store := store2.NewStorage(conn)
	// Create a user
	store2.Seed(store)
}
