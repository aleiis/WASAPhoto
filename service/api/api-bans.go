package api

import (
	"encoding/json"
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type Ban struct {
	BanIssuer  int64 `json:"ban_issuer"`
	BannedUser int64 `json:"banned_user"`
}

func (rt *_router) banUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	if params, err := checkIds(ps.ByName("userId")); err != nil {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	} else {
		userId = params[0]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Decode the user ID to ban from the body of the request
	var ban Ban
	if err := json.NewDecoder(r.Body).Decode(&ban); err != nil {
		http.Error(w, "Error decoding the request body.", http.StatusBadRequest)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, ban.BannedUser} {
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

	// Check if the user ID from the URL is the same as the one in the request body
	if userId != ban.BanIssuer {
		http.Error(w, "User ID mismatch. The user ID in the URL must be the same as the one in the request body.", http.StatusBadRequest)
		return
	}

	// Check if the user is trying to ban itself
	if ban.BanIssuer == ban.BannedUser {
		http.Error(w, "Can't ban yourself.", http.StatusBadRequest)
		return
	}

	// Check if the ban already exists
	if exists, err := rt.db.BanExists(ban.BanIssuer, ban.BannedUser); err != nil {
		ctx.Logger.WithError(err).Error("can't check if ban already exists")
		http.Error(w, "Error checking if the ban exists.", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "Already banned the user.", http.StatusBadRequest)
		return
	}

	// Try to ban the user
	err := rt.db.CreateBan(ban.BanIssuer, ban.BannedUser)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't ban the user")
		http.Error(w, "Error banning the user.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ban)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the ban")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) unbanUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, bannedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("bannedId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		bannedId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, bannedId} {
		exists, err := rt.db.UserExists(id)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't check if the user exists")
			http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
			return
		} else if !exists {
			http.Error(w, "User not found.", http.StatusNotFound)
			return
		}
	}

	// Check if the user was already banned
	if exists, err := rt.db.BanExists(userId, bannedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		http.Error(w, "Error checking if the ban exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Ban not found. Can't unban the user.", http.StatusNotFound)
		return
	}

	// Try to unban the user
	err := rt.db.DeleteBan(userId, bannedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		http.Error(w, "Error unbanning the user.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (rt *_router) checkBanHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, bannedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("bannedId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId, bannedId = params[0], params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the user has banned the user to check
	exists, err := rt.db.BanExists(userId, bannedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the ban exists")
		http.Error(w, "Error checking if the ban exists.", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Ban not found.", http.StatusNotFound)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Ban{BanIssuer: userId, BannedUser: bannedId})
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the ban")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}
