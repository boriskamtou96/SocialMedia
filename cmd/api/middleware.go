package main

import (
	"SocialMedia/internal/store"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *Application) BasicAuthMiddleware() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header from the request
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedError(w, r, fmt.Errorf("missing Authorization header"))
				return
			}
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
				app.unauthorizedError(w, r, fmt.Errorf("invalid Authorization header format"))
				return
			}
			// decode it -> get the username and password
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid base64 encoding in Authorization header"))
				return
			}
			// check the credentials against the config
			username := app.config.auth.basic.user
			password := app.config.auth.basic.password

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 {
				app.unauthorizedError(w, r, fmt.Errorf("invalid credentials format"))
				return
			}

			if creds[0] != username || creds[1] != password {
				app.unauthorizedError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})

	}
}

func (app *Application) AuthTokenMiddleware() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// read the auth header from the request
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedError(w, r, fmt.Errorf("missing Authorization header"))
				return
			}
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				app.unauthorizedError(w, r, fmt.Errorf("invalid Authorization header format"))
				return
			}

			token := parts[1]
			jwtToken, err := app.authenticator.ValidateToken(token)
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid token: %v", err))
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok || claims == nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid token claims"))
				return
			}
			subValue, ok := claims["sub"]
			if !ok {
				app.unauthorizedError(w, r, fmt.Errorf("missing user ID in token"))
				return
			}
			userID, err := strconv.ParseInt(fmt.Sprintf("%.f", subValue), 10, 64)
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid user ID in token"))
				return
			}
			// set the user ID in the request context
			user, err := app.getUser(r.Context(), userID)
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("user not found"))
				return
			}
			if !user.IsActive {
				app.unauthorizedError(w, r, fmt.Errorf("user is not active"))
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userCtx, user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// checkPostOwnerShip lets the post owner through, otherwise requires at least the given role.
func (app *Application) checkPostOwnerShip(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)
		if user != nil && post != nil && post.UserID == user.ID {
			next(w, r)
			return
		}

		app.RequireRoleMiddleware(role)(next).ServeHTTP(w, r)
	}
}

// RequireRoleMiddleware requires the authenticated user to have at least the given role.
func (app *Application) RequireRoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := getUserFromContext(r)
			if user == nil {
				app.unauthorizedError(w, r, fmt.Errorf("user not found in context"))
				return
			}

			allowed, err := app.checkRolePrecedence(r.Context(), user, role)
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}

			if !allowed {
				app.forbiddenResponse(w, r, fmt.Errorf("forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *Application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, fmt.Errorf("role not found: %v", err)
	}

	// Implement role precedence logic here
	return user.Role.Level >= role.Level, nil
}

func (app *Application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	user, err := app.cacheStore.Users.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching user from cache: %w", err)
	}
	if user == nil {
		user, err = app.store.Users.GetById(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		if cacheErr := app.cacheStore.Users.Set(ctx, user); cacheErr != nil {
			app.logger.Warnw("failed to cache user", "userID", userID, "error", cacheErr)
		}
	}

	return user, nil
}

func (app *Application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
