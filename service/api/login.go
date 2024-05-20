package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
)

// doLoginHandler is an HTTP handler that returns the
func (rt *_router) doLoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Decode the username from the body of the request
	var username string
	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		ctx.Logger.WithError(err).Error("can't decode the username", username)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username has the correct format
	if !checkUsername(username) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to get the user ID from the database
	var id int64
	id, err := rt.db.GetUserId(username)
	if err != nil {
		// If the user is not found, create a new user
		if errors.Is(err, database.ErrUserNotFound) {
			id, err = rt.db.CreateUser(username)
			if err != nil {
				ctx.Logger.WithError(err).Error("can't create the user")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			ctx.Logger.WithError(err).Error("can't get the user ID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Get the bearer token of the user
	bearer, err := getBearer(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the bearer token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, `%s`, bearer)
}
