package main

import (
	"SocialMedia/internal/store"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

type CreateUserPayload struct {
	Username string `json:"username" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}

// createUserHandler godoc
//
//	@Summary		Create a user
//	@Description	Registers a new account. The password is never returned.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserPayload	true	"Account details"
//	@Success		201		{object}	Envelope{data=store.User}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users [post]
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

	if err := JsonResponse(w, http.StatusCreated, u); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getUsersHandler godoc
//
//	@Summary	List users
//	@Tags		users
//	@Produce	json
//	@Success	200	{object}	Envelope{data=[]store.User}
//	@Failure	500	{object}	ErrorResponse
//	@Router		/users [get]
func (app *Application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := app.store.Users.GetUsers(ctx)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := JsonResponse(w, http.StatusOK, users); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getUserByIdHandler godoc
//
//	@Summary	Fetch a user
//	@Tags		users
//	@Produce	json
//	@Param		userID	path		int	true	"User ID"
//	@Success	200		{object}	Envelope{data=store.User}
//	@Failure	400		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/users/{userID} [get]
func (app *Application) getUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := JsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

// followUserHandler godoc
//
//	@Summary		Follow a user
//	@Description	The user in the path starts following the user in the body.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	int			true	"Follower user ID"
//	@Param			payload	body	FollowUser	true	"User to follow"
//	@Success		204		"No content"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/{userID}/follow [put]
func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)

	var payload FollowUser
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	userID := payload.UserID

	log.Println("User Id:", userID)
	// Check if the user to be followed exists
	_, err := app.store.Users.GetById(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.badRequestError(w, r, errors.New("user to follow does not exist"))
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.store.Followers.Follow(ctx, followerUser.ID, userID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := JsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// unFollowUserHandler godoc
//
//	@Summary	Unfollow a user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		userID	path	int			true	"Follower user ID"
//	@Param		payload	body	FollowUser	true	"User to unfollow"
//	@Success	204		"No content"
//	@Failure	400		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/users/{userID}/unfollow [put]
func (app *Application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unFollowerUser := getUserFromContext(r)

	var payload FollowUser
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	userID := payload.UserID
	if err := app.store.Followers.UnFollow(ctx, unFollowerUser.ID, userID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := JsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *Application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtx).(*store.User)
	if !ok {
		return nil
	}
	return user
}
