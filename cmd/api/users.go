package main

import (
	"SocialMedia/internal/store"
	"net/http"
	"time"
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
