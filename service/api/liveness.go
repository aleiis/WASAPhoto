package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// liveness is an HTTP handler that checks the API server status. If the server cannot serve requests (e.g., some
// resources are not ready), this should reply with HTTP Status 500. Otherwise, with HTTP Status 200
// noinspection GoUnusedParameter
func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Check if the database is ready
	if err := rt.db.Ping(); err != nil {
		http.Error(w, "Database is not ready", http.StatusInternalServerError)
		return
	}

	// If everything is ready, reply with HTTP Status 200
	w.WriteHeader(http.StatusOK)
}
