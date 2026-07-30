package main

import (
	"log"
	"net/http"
)

func (app *Application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s\n", err)
	ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
}

func (app *Application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found: %s\n", err)
	ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
}

func (app *Application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %s\n", err)
	ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
}
