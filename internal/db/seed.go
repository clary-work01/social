package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/chainflow/chainflow-api/internal/store"
)

var usernames = []string{
	"alice_wonder",
	"bob_the_builder",
	"charlie_brown",
	"diana_prince",
	"evan_storm",
	"fiona_green",
	"george_banks",
	"hannah_rose",
	"ivan_drago",
	"julia_fox",
	"kevin_hart",
	"laura_palmer",
	"mike_tyson",
	"nina_simone",
	"oscar_wilde",
	"paula_white",
	"quinn_sharp",
	"rachel_green",
	"steve_rogers",
	"tina_turner",
}

var postTitles = []string{
	"Getting Started with Go Backend Development",
	"Understanding Goroutines and Channels",
	"Building REST APIs with Chi Router",
	"PostgreSQL Best Practices for Developers",
	"Redis Caching Strategies Explained",
	"How to Handle Errors in Go",
	"JWT Authentication from Scratch",
	"Database Migrations with golang-migrate",
	"Writing Clean Code in Go",
	"Concurrency Patterns You Should Know",
	"Docker Compose for Local Development",
	"Middleware Design in HTTP Servers",
	"Optimizing SQL Queries with Indexes",
	"Graceful Shutdown in Go Services",
	"Introduction to Context Package",
	"Unit Testing Go Applications",
	"From Zero to REST API in Go",
	"Understanding Go Interfaces",
	"Rate Limiting Your API",
	"Structured Logging with slog",
}

var postContents = []string{
	"Go is a statically typed language designed for simplicity and performance. In this post, we explore the basics of setting up a Go backend project from scratch.",
	"Goroutines are lightweight threads managed by the Go runtime. Combined with channels, they make concurrent programming intuitive and safe.",
	"Chi is a lightweight router built on top of net/http. It supports middleware chaining, URL parameters, and route grouping out of the box.",
	"Choosing the right data types, using indexes wisely, and writing efficient queries are the foundations of working with PostgreSQL in production.",
	"Redis supports multiple caching strategies including cache-aside, write-through, and write-behind. Choosing the right one depends on your consistency requirements.",
	"Go encourages explicit error handling. Wrapping errors with fmt.Errorf and using errors.Is or errors.As gives you full control over the error chain.",
	"JSON Web Tokens provide a stateless authentication mechanism. We cover signing, verifying, and storing claims securely in your Go API.",
	"Database migrations let you version control your schema changes. golang-migrate supports both up and down migrations with multiple database drivers.",
	"Clean code in Go means short functions, meaningful names, and avoiding unnecessary abstractions. The standard library itself is a great reference.",
	"From WaitGroups to mutexes and semaphores, Go provides powerful primitives for managing concurrent workloads safely and efficiently.",
	"Docker Compose simplifies running multi-container applications locally. Define your API, database, and cache in a single docker-compose.yml file.",
	"Middleware in Go wraps HTTP handlers to add cross-cutting concerns like logging, authentication, and rate limiting without touching business logic.",
	"Indexes dramatically speed up SELECT queries but slow down writes. Understanding B-tree indexes and query plans will help you make better decisions.",
	"A graceful shutdown ensures in-flight requests are completed before the server exits. We use os.Signal and context cancellation to implement it.",
	"The context package lets you carry deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines.",
	"Good tests give you confidence to refactor. We cover table-driven tests, mocking interfaces, and using testify to write readable assertions.",
	"In this guide we build a fully functional REST API in Go from zero: routing, middleware, database access, and structured error responses.",
	"Interfaces in Go are satisfied implicitly. This enables powerful patterns like dependency injection and makes your code easier to test.",
	"Rate limiting protects your API from abuse. We implement a token bucket algorithm using Redis to share state across multiple server instances.",
	"The slog package introduced in Go 1.21 provides structured logging with levels, attributes, and pluggable handlers for production observability.",
}

var postTags = []string{
	"golang",
	"backend",
	"postgresql",
	"redis",
	"docker",
	"api",
	"database",
	"concurrency",
	"middleware",
	"authentication",
	"testing",
	"performance",
	"sql",
	"microservices",
	"devops",
	"security",
	"caching",
	"logging",
	"deployment",
	"opensource",
}

var commentContents = []string{
	"Great post, really helped me understand the concept!",
	"I was struggling with this for days, thanks for the clear explanation.",
	"Could you write a follow-up post on this topic?",
	"This is exactly what I was looking for, bookmarked!",
	"I tried this approach in my project and it worked perfectly.",
	"Nice writeup, but I think there's a typo in the second code snippet.",
	"Very well explained, even a beginner can follow along.",
	"I disagree with the approach here, using X would be more efficient.",
	"Just what I needed for my side project, appreciate it!",
	"The section on error handling was especially useful.",
	"Would love to see a video version of this tutorial.",
	"I've been doing it the wrong way this whole time, thanks for the correction.",
	"Short and to the point, no fluff. Exactly how tutorials should be.",
	"The code examples are clean and easy to understand.",
	"This saved me hours of debugging, thank you so much.",
	"Can you share the full source code on GitHub?",
	"I shared this with my team, everyone found it helpful.",
	"The explanation of the underlying mechanism is really insightful.",
	"Looking forward to more posts like this one.",
	"Finally a post that explains this without overcomplicating things.",
}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.User.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Seeding Users Error:", err)
			return
		}
	}

	tx.Commit()

	posts := generatePosts(100, users)
	for _, post := range posts {
		if err := store.Post.Create(ctx, post); err != nil {
			log.Println("Seeding Posts Error:", err)
			return
		}
	}

	comments := generateComments(100, users, posts)
	for _, comment := range comments {
		if err := store.Comment.Create(ctx, comment); err != nil {
			log.Println("Seeding Comments Error:", err)
			return
		}
	}

	log.Println("Seeding Complete !")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			UserName: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Role: store.Role{
				Name: "user",
			},
		}

		err := users[i].Password.Set("123123")
		log.Println("generate Users Error:", err)
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	for i := 0; i < num; i++ {
		posts[i] = &store.Post{
			Title:   postTitles[rand.Intn(len(postTitles))] + fmt.Sprintf("%d", i),
			Content: postContents[rand.Intn(len(postContents))] + fmt.Sprintf("%d", i),
			UserID:  users[rand.Intn(len(users))].ID,
			Tags: []string{
				postTags[rand.Intn(len(postTags))],
				postTags[rand.Intn(len(postTags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, num)

	for i := 0; i < num; i++ {
		comments[i] = &store.Comment{
			UserID:  users[rand.Intn(len(users))].ID,
			PostID:  posts[rand.Intn(len(posts))].ID,
			Content: commentContents[rand.Intn(len(commentContents))] + fmt.Sprintf("%d", i),
		}
	}

	return comments
}
