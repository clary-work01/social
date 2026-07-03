package main

import (
	"log"
	"time"

	"github.com/chainflow/chainflow-api/internal/auth"
	"github.com/chainflow/chainflow-api/internal/cache"
	"github.com/chainflow/chainflow-api/internal/db"
	"github.com/chainflow/chainflow-api/internal/env"
	"github.com/chainflow/chainflow-api/internal/mailer"
	"github.com/chainflow/chainflow-api/internal/ratelimiter"
	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const version = "0.0.2"

//	@title			Chainflow API
//	@description	API for Chainflow.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	godotenv.Load()

	cfg := config{
		port:        env.GetEnvString("SERVER_PORT", ":8080"),
		env:         env.GetEnvString("ENV", "development"),
		externalURL: env.GetEnvString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetEnvString("FRONTEND_URL", "http://localhost:4000"),
		db: dbConfig{
			dsn:          env.GetEnvString("DB_DSN", "postgres://admin:admin123@localhost/chainflow_db?sslmode=disable"),
			maxOpenConns: env.GetEnvInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetEnvInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetEnvString("DB_MAX_IDLE_TIME", "15m"),
		},
		redis: redisConfig{
			addr:     env.GetEnvString("REDIS_ADDR", "localhost:6379"),
			password: env.GetEnvString("REDIS_PW", ""),
			db:       env.GetEnvInt("REDIS_DB", 0),
			enabled:  env.GetEnvBool("REDIS_ENABLED", false),
		},
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetEnvString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetEnvString("SENDGRID_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicAuthConfig{
				user: env.GetEnvString("AUTH_BASIC_USER", "admin"),
				pass: env.GetEnvString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetEnvString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3, // 3 days for testing
				iss:    "chainflow",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetEnvInt("RATELIMITER_REQUEST_PER_TIME_FRAME", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetEnvBool("RATELIMITER_ENABLED", true),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync() // flushes buffer, if any

	// Database
	db, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// Redis
	var rdb *redis.Client
	if cfg.redis.enabled {
		rdb = cache.NewRedisClient(
			cfg.redis.addr,
			cfg.redis.password,
			cfg.redis.db,
		)
		logger.Info("redis connection established")
	}

	// Mailer
	sendgridMailer := mailer.NewSendGrid(
		cfg.mail.sendGrid.apiKey,
		cfg.mail.fromEmail,
	)

	// Authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(
		cfg.auth.token.secret,
		cfg.auth.token.iss,
		cfg.auth.token.iss, // we are the only consumer of this token for now
	)

	// Rate Limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	app := &application{
		config:        cfg,
		logger:        logger,
		store:         store.NewStorage(db),
		cache:         cache.NewRedisStorage(rdb),
		mailer:        sendgridMailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	router := app.mount()
	if err := app.run(router); err != nil {
		log.Fatal(err)
	}
}
