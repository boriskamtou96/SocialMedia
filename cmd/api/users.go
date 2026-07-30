package main

import (
	"SocialMedia/internal/store"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type CreateUserPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload CreateUserPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	u := &store.User{
		Username:  payload.Username,
		Email:     payload.Email,
		Password:  payload.Password,
		CreatedAt: time.Now().String(),
	}

	err := app.store.Users.Create(ctx, u)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	if err := WriteJSON(w, http.StatusCreated, u); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}
}

func (app *Application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.store.Users.GetUsers(ctx)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	if err := WriteJSON(w, http.StatusOK, users); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}
}

func (app *Application) getUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIdParam := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(userIdParam, 10, 64)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID format")
		return
	}

	user, err := app.store.Users.GetById(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}

		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	if user == nil {
		ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	if err := WriteJSON(w, http.StatusOK, user); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

}
