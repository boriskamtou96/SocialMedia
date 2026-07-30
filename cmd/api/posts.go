package main

import (
	"SocialMedia/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,min=1,max=255"`
	Content string   `json:"content" validate:"required,min=20,max=1000"`
	Tags    []string `json:"tags" validate:"required,min=1,max=10"`
}

func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := WriteJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(postIDParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, errors.New("invalid post ID format"))
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := WriteJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
