package main

import (
	"SocialMedia/internal/mailer"
	"SocialMedia/internal/store"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type userKey string

const userCtx userKey = "user"

type CreateUserPayload struct {
	Username string `json:"username" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
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
		CreatedAt: time.Now().String(),
	}

	err := u.Password.Set(payload.Password)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	plainToken := uuid.New().String() // Generate a new UUID for the invitation token
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err = app.store.Users.CreateAndInvite(ctx, u, hashToken, app.config.mail.exp)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDuplicateUsername):
			app.badRequestError(w, r, err)
		case errors.Is(err, store.ErrDuplicateEmail):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	fmt.Println("Activation Token (send this to the user via email):", plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      u.Username,
		ActivationURL: fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken),
	}
	status, err := app.mailer.Send(mailer.UserInvitationTemplate, u.Username, u.Email, vars, false)
	if err != nil {
		app.logger.Errorw("failed to send activation email", "error", err, "userID", u.ID, "email", u.Email)
		// rollback user creation if email sending fails
		if rollbackErr := app.store.Users.DeleteById(ctx, u.ID); rollbackErr != nil {
			app.logger.Errorw("failed to rollback user creation", "error", rollbackErr, "userID", u.ID)
		}
		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("activation email sent with status", "userID", u.ID, "email", u.Email, "status", status)

	if err := JsonResponse(w, http.StatusCreated, u); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// activateUserHandler godoc
//
//	@Summary	Activate a user
//	@Description	Activates a user account using the provided activation token.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			activationToken	path	string	true	"Activation Token"
//	@Success		204				"No content"
//	@Failure		400				{object}	ErrorResponse
//	@Failure		404				{object}	ErrorResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/users/activate/{token} [put]
func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	if token == "" {
		app.badRequestError(w, r, errors.New("activation token is required"))
		return
	}

	err := app.store.Users.Activate(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// 204 must not carry a body, so no JsonResponse here
	w.WriteHeader(http.StatusNoContent)
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
	if followerUser == nil {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, errors.New("invalid user ID format"))
		return
	}

	var payload FollowUser
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	userID := payload.UserID

	// Check if the user to be followed exists
	_, err = app.store.Users.GetById(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.badRequestError(w, r, errors.New("user to follow does not exist"))
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.store.Followers.Follow(ctx, followerUser.ID, followedID); err != nil {
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
	followerUser := getUserFromContext(r)
	if followerUser == nil {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	unFollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, errors.New("invalid user ID format"))
		return
	}

	var payload FollowUser
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Followers.UnFollow(ctx, followerUser.ID, unFollowedID); err != nil {
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

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// createTokenHandler godoc
//
//	@Summary		Create an authentication token
//	@Description	Generates a new authentication token for a user.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/authentication/token [post]
func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	// fetch the user (check if the user exists) from the payload
	user, err := app.store.Users.GetUserByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedError(w, r, errors.New("invalid credentials"))
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedError(w, r, errors.New("invalid credentials"))
		return
	}

	// generate the token -> add claims
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "SocialMediaApp",
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := JsonResponse(w, http.StatusOK, map[string]string{"token": token}); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
