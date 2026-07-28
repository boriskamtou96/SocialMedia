package main

import (
	"SocialMedia/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
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
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}

	if err := WriteJSON(w, http.StatusCreated, post); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}
}

func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(postIDParam, 10, 64)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid post ID format")
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		default:
			ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		}
		return
	}

	if err := WriteJSON(w, http.StatusOK, post); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
		return
	}
}
