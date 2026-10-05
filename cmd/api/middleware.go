package main

import (
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
			// Parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
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
			// Parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				app.unauthorizedError(w, r, fmt.Errorf("invalid Authorization header format"))
				return
			}

			token := parts[1]
			// validate the token
			jwtToken, err := app.authenticator.ValidateToken(token)
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid token: %v", err))
				return
			}

			claims, _ := jwtToken.Claims.(jwt.MapClaims)
			userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
			if err != nil {
				app.unauthorizedError(w, r, fmt.Errorf("invalid user ID in token"))
				return
			}
			// set the user ID in the request context
			user, err := app.store.Users.GetById(r.Context(), userID)
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
