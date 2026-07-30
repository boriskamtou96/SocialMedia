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
	Username string `json:"username" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}

func (app *Application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload CreateUserPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
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
		app.internalServerError(w, r, err)
		return
	}

	if err := WriteJSON(w, http.StatusCreated, u); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.store.Users.GetUsers(ctx)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, users); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) getUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIdParam := chi.URLParam(r, "userID")
	userID, err := strconv.ParseInt(userIdParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, errors.New("invalid user ID format"))
		return
	}

	user, err := app.store.Users.GetById(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}

		app.internalServerError(w, r, err)
		return
	}

	if user == nil {
		app.notFoundError(w, r, errors.New("user not found"))
		return
	}

	if err := WriteJSON(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
