package main

import (
	"net/http"
)

func (app *Application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("internal server error: %s", err)
	ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
}

func (app *Application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("not found: %s", err)
	ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
}

func (app *Application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("bad request: %s", err)
	ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
}

func (app *Application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("unauthorized: %s", err)
	ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
}

func (app *Application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("forbidden: %s", err)
	ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
}

func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded: %s", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "rate limit exceeded, retry after: "+retryAfter)
}
