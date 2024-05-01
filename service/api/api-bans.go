package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) banUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	fmt.Println("banUserHandler")
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

	// Decode the user ID to ban from the body of the request
	var bannedUserId int64
	if err := json.NewDecoder(r.Body).Decode(&bannedUserId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user ID to ban is valid
	if exists, err := rt.db.UserIdExists(bannedUserId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists || bannedUserId == userId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the ban already exists
	exists, err = rt.db.BanExists(userId, bannedUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if ban already exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the banned user was following the user
	exists, err = rt.db.FollowExists(bannedUserId, userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		rt.db.DeleteFollow(bannedUserId, userId)
	}

	// Try to ban the user
	err = rt.db.CreateBan(userId, bannedUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't ban the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) unbanUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	strUserId := ps.ByName("userId")
	strBannedUserId := ps.ByName("bannedId")

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

	// Check if the banned user ID is a valid int64
	bannedUserId, err := strconv.ParseInt(strBannedUserId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user ID to unban is valid
	if exists, err := rt.db.UserIdExists(bannedUserId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists || bannedUserId == userId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user was already banned
	exists, err = rt.db.BanExists(userId, bannedUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Try to unfollow the user
	err = rt.db.DeleteBan(userId, bannedUserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
