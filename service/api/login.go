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

func (rt *_router) doLoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Decode the username from the body of the request
	var username string
	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username contains at least 3 characters and no more than 16
	if len(username) < 3 || len(username) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username contains only alphanumeric characters
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `%d`, id)
}
