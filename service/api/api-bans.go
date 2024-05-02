package api

import (
	"encoding/json"
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (rt *_router) banUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	if params, err := checkIds(ps.ByName("userId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId = params[0]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Decode the user ID to ban from the body of the request
	var bannedId int64
	if err := json.NewDecoder(r.Body).Decode(&bannedId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, bannedId} {
		exists, err := rt.db.UserExists(id)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't check if the user exists")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Check if the ban already exists
	if exists, err := rt.db.BanExists(userId, bannedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if ban already exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user is trying to ban itself
	if userId == bannedId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to ban the user
	err := rt.db.CreateBan(userId, bannedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't ban the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) unbanUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, bannedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("bannedId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		bannedId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, bannedId} {
		exists, err := rt.db.UserExists(id)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't check if the user exists")
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// Check if the user was already banned
	if exists, err := rt.db.BanExists(userId, bannedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Try to unban the user
	err := rt.db.DeleteBan(userId, bannedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
