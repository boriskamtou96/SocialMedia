package main

import (
	"SocialMedia/internal/store"
	"context"
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

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"required,min=1,max=255"`
	Content *string `json:"content" validate:"required,min=20,max=1000"`
}

// createPostHandler godoc
//
//	@Summary	Create a post
//	@Tags		posts
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		CreatePostPayload	true	"Post to create"
//	@Success	201		{object}	Envelope{data=store.Post}
//	@Failure	400		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/posts [post]
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

	if err := JsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPostHandler godoc
//
//	@Summary		Fetch a post
//	@Description	Returns the post along with its comments.
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	Envelope{data=store.Post}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID} [get]
func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := JsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// deletePostHandler godoc
//
//	@Summary	Delete a post
//	@Tags		posts
//	@Produce	json
//	@Param		postID	path		int	true	"Post ID"
//	@Success	200		{object}	Envelope{data=map[string]string}
//	@Failure	400		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/posts/{postID} [delete]
func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postIdParam := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(postIdParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	err = app.store.Posts.DeletePostById(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)

		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := JsonResponse(w, http.StatusOK, map[string]string{"message": "Post deleted successfully"}); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// updatePostHandler godoc
//
//	@Summary		Update a post
//	@Description	Only the fields present in the body are modified.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Fields to update"
//	@Success		200		{object}	Envelope{data=store.Post}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/posts/{postID} [patch]
func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	err := app.store.Posts.UpdatePostById(r.Context(), post)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := JsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		postIdParam := chi.URLParam(r, "postID")
		postID, err := strconv.ParseInt(postIdParam, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

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

		ctx = context.WithValue(r.Context(), "post", post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value("post").(*store.Post)
	return post
}
