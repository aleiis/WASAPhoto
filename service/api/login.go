package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
)

// doLoginHandler is an HTTP handler that returns the
func (rt *_router) doLoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Decode the username from the body of the request
	var username Username
	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		http.Error(w, "Invalid request body. The username could not be decoded.", http.StatusBadRequest)
		return
	}

	// Check if the username has the correct format
	if !checkUsernameFormat(username.Username) {
		http.Error(w, "Invalid username. The username must be a string with 3 to 16 alphanumeric characters.", http.StatusBadRequest)
		return
	}

	// Try to get the user ID from the database
	var id int64
	id, err := rt.db.GetUserId(username.Username)
	if err != nil {
		// If the user is not found, create a new user
		if errors.Is(err, database.ErrUserNotFound) {
			id, err = rt.db.CreateUser(username.Username)
			if err != nil {
				ctx.Logger.WithError(err).Error("can't create the user")
				http.Error(w, "Can't create the user.", http.StatusInternalServerError)
				return
			}
		} else {
			ctx.Logger.WithError(err).Error("can't get the user ID")
			http.Error(w, "Can't get the user ID.", http.StatusInternalServerError)
			return
		}
	}

	// Get the bearer token of the user
	bearer, err := getBearer(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the bearer token")
		http.Error(w, "Can't get the bearer token.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(bearer); err != nil {
		ctx.Logger.WithError(err).Error("can't encode the bearer token")
		http.Error(w, "Can't encode the bearer token.", http.StatusInternalServerError)
		return
	}
}
