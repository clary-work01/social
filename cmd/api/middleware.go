package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

// func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// read the auth header
// 			authHeader := r.Header.Get("Authorization")
// 			if authHeader == "" {
// 				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is missing"))
// 				return
// 			}

// 			// parse it -> get the base64
// 			parts := strings.Split(authHeader, " ")
// 			if len(parts) != 2 || parts[0] != "Basic" {
// 				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header malformed"))
// 				return
// 			}

// 			// decode it
// 			decoded, err := base64.StdEncoding.DecodeString(parts[1])
// 			if err != nil {
// 				app.unauthorizedBasicError(w, r, err)
// 				return
// 			}

// 			// check the credentials
// 			user := app.config.auth.basic.user
// 			pass := app.config.auth.basic.pass

// 			creds := strings.SplitN(string(decoded), ":", 2)
// 			if len(creds) != 2 || creds[0] != user || creds[1] != pass {
// 				app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
// 				return
// 			}

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

func (app *application) tokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read the auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		// parse it -> get the base64
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		// parse the claims, grab the userID and fetch the user from DB
		jwtToken, err := app.authenticator.ValidateToken(parts[1])
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		sub, ok := claims["sub"].(float64) // JWT decode 後數字都是 float64
		if !ok {
			app.unauthorizedError(w, r, fmt.Errorf("invalid sub claim"))
			return
		}
		userID := int64(sub)

		ctx := r.Context()
		user, err := app.getUser(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				// 不回傳404 怕有人用此API測試某email是否存在
				app.unauthorizedError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		// inject the authenticated user into the context
		ctx = context.WithValue(ctx, userCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnershipMiddleware(requiredRoleName string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post := app.getPostFromCtx(r)
		user := app.getUserFromCtx(r)

		// if it is the user's post
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		// role precedence check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRoleName)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenError(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, requiredRoleName string) (bool, error) {
	requiredRole, err := app.store.Role.GetByName(ctx, requiredRoleName)
	if err != nil {
		return false, nil
	}

	app.logger.Info(user, requiredRoleName)
	return user.Role.Level >= requiredRole.Level, nil
}

// 這段可以做Cache
func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.config.redis.enabled {
		return app.store.User.GetByID(ctx, userID)
	}

	// app.logger.Infow("cache hit", "key", "user", "id", userID)

	user, err := app.cache.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// user不在cache, 才去DB找
	if user == nil {
		// app.logger.Infow("fetching from DB", "id", userID)
		user, err = app.store.User.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		// DB找到後Set進cache
		if err := app.cache.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// app.logger.Infow("rate limiter check", "remoteAddr", r.RemoteAddr)

		if app.config.rateLimiter.Enabled {
			// 用 net.SplitHostPort 把 port 拆掉, 因為每次請求的 port 都不一樣
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// 萬一 RemoteAddr 格式不含 port（理論上不太會發生），就整串當 fallback
				ip = r.RemoteAddr
			}
			if allow, retryAfter := app.rateLimiter.Allow(ip); !allow {
				app.rateLimitExceededError(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
