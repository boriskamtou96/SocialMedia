package main

import (
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	data := map[string]string{
		"status":  "available",
		"env":     "dev",
		"version": "1",
	}
	WriteJSON(w, http.StatusOK, data)
}
