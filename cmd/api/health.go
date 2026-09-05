package main

import (
	"net/http"
)

// healthCheckHandler godoc
//
//	@Summary		Service health
//	@Description	Returns the status, environment and version of the API.
//	@Tags			ops
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (app *Application) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	data := map[string]string{
		"status":  "available",
		"env":     "dev",
		"version": "1",
	}
	WriteJSON(w, http.StatusOK, data)
}
