package store

import (
	"context"
	"log"
	"math/rand"

	"github.com/bxcodec/faker/v4"
)

func Seed(store Storage) {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		err := store.Users.Create(ctx, user)
		if err != nil {
			log.Println("error creating user:", err)
			return
		}
	}

	posts := generatePosts(200, users)
	for _, post := range posts {
		err := store.Posts.Create(ctx, post)
		if err != nil {
			log.Println("error creating post:", err)
			return
		}
	}

	comments := generateComments(200, posts, users)
	for _, comment := range comments {
		err := store.Comments.Create(ctx, &comment)
		if err != nil {
			log.Println("error creating comment:", err)
			return
		}
	}

	log.Println("Seeding completed successfully")
}

func generateUsers(count int) []*User {
	users := make([]*User, count)

	for i := 0; i < count; i++ {
		users[i] = &User{
			Username: "",
			Email:    "",
			Password: "",
		}
		users[i].Username = faker.Username()
		users[i].Email = faker.Email()
		users[i].Password = faker.Password()
	}
	return users

}

func generatePosts(count int, users []*User) []*Post {
	posts := make([]*Post, count)

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		posts[i] = &Post{
			Title:   "",
			Content: "",
			UserID:  0,
		}
		posts[i].Title = faker.Sentence()
		posts[i].Content = faker.Paragraph()
		posts[i].Tags = []string{faker.Word(), faker.Word(), faker.Word()}
		posts[i].Comments = []Comment{}
		posts[i].UserID = user.ID
	}
	return posts

}

func generateComments(count int, posts []*Post, users []*User) []Comment {
	comments := make([]Comment, count)

	for i := 0; i < count; i++ {
		post := posts[rand.Intn(len(posts))]
		user := users[rand.Intn(len(users))]
		comments[i] = Comment{
			PostID:  post.ID,
			UserID:  user.ID,
			Content: faker.Sentence(),
		}
	}
	return comments
}
