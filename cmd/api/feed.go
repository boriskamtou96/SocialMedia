package main

import (
	"SocialMedia/internal/store"
	"net/http"
)

// getUserFeedHandler godoc
//
//	@Summary		User feed
//	@Description	Posts from the accounts the user follows, with their comment count.
//	@Tags			feed
//	@Produce		json
//	@Param			limit	query		int			false	"Page size, 1 to 20"	default(20)
//	@Param			offset	query		int			false	"Rows to skip"			default(0)
//	@Param			sort	query		string		false	"Sort order"			Enums(asc, desc)	default(desc)
//	@Param			tags	query		[]string	false	"Filter by tags, up to 5"
//	@Param			search	query		string		false	"Full text search on title and content"
//	@Param			since	query		string		false	"Lower bound date, YYYY-MM-DD"
//	@Param			until	query		string		false	"Upper bound date, YYYY-MM-DD"
//	@Success		200		{object}	Envelope{data=[]store.PostWithMetaData}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/users/feed [get]
func (app *Application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	posts, err := app.store.Posts.GetUserFeed(ctx, int64(110), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := JsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
