package main

import (
	"log"
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Println("Error writing health check response:", err)
	}
}
