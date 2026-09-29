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
