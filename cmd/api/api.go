package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chainflow/chainflow-api/docs" // This is required to generate swagger docs
	"github.com/chainflow/chainflow-api/internal/auth"
	"github.com/chainflow/chainflow-api/internal/cache"
	"github.com/chainflow/chainflow-api/internal/env"
	"github.com/chainflow/chainflow-api/internal/mailer"
	"github.com/chainflow/chainflow-api/internal/ratelimiter"
	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type config struct {
	port        string
	env         string
	externalURL string
	frontendURL string
	db          dbConfig
	redis       redisConfig
	mail        mailConfig
	auth        authConfig
	rateLimiter ratelimiter.Config
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type sendGridConfig struct {
	apiKey string
}

type authConfig struct {
	basic basicAuthConfig
	token tokenConfig
}

type basicAuthConfig struct {
	user string
	pass string
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type redisConfig struct {
	addr    string
	db      int
	enabled bool // redis can be optional
}

type application struct {
	config        config
	store         store.Storage
	cache         cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetEnvString("CORS_ALLOWED_ORIGIN", "http://localhost:5174")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.rateLimiterMiddleware)

	r.Route("/v1", func(r chi.Router) {
		// r.With(app.basicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.port)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.tokenAuthMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPostOwnershipMiddleware("admin", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnershipMiddleware("moderator", app.updatePostHandler))
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler) // Public Route

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.tokenAuthMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unFollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.tokenAuthMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public Routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.externalURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.port,
		Handler:      mux,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	// graceful shutdown
	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Info("server started at:", "port", app.config.port)

	// ListenAndServe 是阻塞的，它會一直跑直到 server 停止。停止時會回傳一個 error
	// 而這個 error 有兩種情況：
	// 1.正常 graceful shutdown（被 srv.Shutdown() 呼叫）http.ErrServerClosed
	// 2.真的出錯（port 被佔用、網路問題等）
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped",
		"addr", app.config.port,
		"env", app.config.env,
	)

	return nil
	// return srv.ListenAndServe()
}
