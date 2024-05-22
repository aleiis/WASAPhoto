package api

import (
	"encoding/json"
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (rt *_router) followUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

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

	// Decode the user ID to follow from the body of the request
	var followedId int64
	if err := json.NewDecoder(r.Body).Decode(&followedId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, followedId} {
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

	// Check if the user is trying to follow itself
	if userId == followedId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the follow already exists
	if exists, err := rt.db.FollowExists(userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user to follow has banned the user
	if exists, err := rt.db.BanExists(followedId, userId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user has banned the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Try to follow the user
	if err := rt.db.CreateFollow(userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't follow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) unfollowUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, followedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("followedId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		followedId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, followedId} {
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

	// Check if the user is following the user to unfollow
	if exists, err := rt.db.FollowExists(userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Try to unfollow the user
	if err := rt.db.DeleteFollow(userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func (rt *_router) checkFollowHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, followedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("followedId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId, followedId = params[0], params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user is following the user to check
	exists, err := rt.db.FollowExists(userId, followedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(404)
	}
}
