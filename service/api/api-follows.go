package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) followUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	strUserId := ps.ByName("userId")

	bearer := r.Header.Get("Authorization")

	if bearer != strUserId {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user ID is a valid int64
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user exists
	exists, err := rt.db.UserIdExists(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Decode the user ID to follow from the body of the request
	var followUserId int64
	if err := json.NewDecoder(r.Body).Decode(&followUserId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user ID to follow is valid
	if exists, err := rt.db.UserIdExists(followUserId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists || followUserId == userId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the follow already exists
	exists, err = rt.db.FollowExists(userId, followUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user to follow has banned the user
	exists, err = rt.db.BanExists(followUserId, userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user has banned the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Try to follow the user
	err = rt.db.CreateFollow(userId, followUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't follow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) unfollowUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	strUserId := ps.ByName("userId")
	strFollowedId := ps.ByName("followedId")

	bearer := r.Header.Get("Authorization")

	if bearer != strUserId {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user ID is a valid int64
	userId, err := strconv.ParseInt(strUserId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user exists
	exists, err := rt.db.UserIdExists(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if the followed ID is a valid int64
	followedId, err := strconv.ParseInt(strFollowedId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user ID to follow is valid
	if exists, err := rt.db.UserIdExists(followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists || followedId == userId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user is following the user to unfollow
	exists, err = rt.db.FollowExists(userId, followedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Try to unfollow the user
	err = rt.db.DeleteFollow(userId, followedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
