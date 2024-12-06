package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Follow struct {
	Follower int64 `json:"follower"`
	Followed int64 `json:"followed"`
}

func (rt *_router) followUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "followUserHandler")
	defer span.End()

	// Get the parameters
	var userId int64
	params, err := checkIds(ps.ByName("userId"))
	if err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	}
	userId = params[0]

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Decode the request body
	var follow Follow
	if err := json.NewDecoder(r.Body).Decode(&follow); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, follow.Followed} {
		exists, err := rt.db.UserExists(otelctx, id)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't check if the user exists")
			http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
			return
		} else if !exists {
			http.Error(w, "User not found.", http.StatusNotFound)
			return
		}
	}

	// Check if the user ID from the URL is the same as the one in the request body
	if userId != follow.Follower {
		http.Error(w, "User ID mismatch. The user ID in the URL must be the same as the one in the request body.", http.StatusBadRequest)
		return
	}

	// Check if the user is trying to follow itself
	if follow.Follower == follow.Followed {
		http.Error(w, "Can't follow yourself.", http.StatusBadRequest)
		return
	}

	// Check if the follow already exists
	if exists, err := rt.db.FollowExists(otelctx, follow.Follower, follow.Followed); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		http.Error(w, "Error checking if the follow exists.", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "Already following the user.", http.StatusBadRequest)
		return
	}

	// Check if the user to follow has banned the user
	if exists, err := rt.db.BanExists(otelctx, follow.Followed, follow.Follower); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user has banned the user")
		http.Error(w, "Error checking if the followed user has banned the follower user.", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "Can't follow the user.", http.StatusForbidden)
		return
	}

	// Try to follow the user
	if err := rt.db.CreateFollow(otelctx, follow.Follower, follow.Followed); err != nil {
		ctx.Logger.WithError(err).Error("can't follow the user")
		http.Error(w, "Error following the user.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(follow)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the follow")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) unfollowUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "unfollowUserHandler")
	defer span.End()

	// Get the parameters
	var userId, followedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("followedId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		followedId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if both user IDs exist
	for _, id := range []int64{userId, followedId} {
		exists, err := rt.db.UserExists(otelctx, id)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't check if the user exists")
			http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
			return
		} else if !exists {
			http.Error(w, "User not found.", http.StatusNotFound)
			return
		}
	}

	// Check if the user is following the user to unfollow
	if exists, err := rt.db.FollowExists(otelctx, userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		http.Error(w, "Error checking if the follow exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Follow not found. Not following the user.", http.StatusNotFound)
		return
	}

	// Try to unfollow the user
	if err := rt.db.DeleteFollow(otelctx, userId, followedId); err != nil {
		ctx.Logger.WithError(err).Error("can't unfollow the user")
		http.Error(w, "Error unfollowing the user.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (rt *_router) checkFollowHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "checkFollowHandler")
	defer span.End()

	// Get the parameters
	var userId, followedId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("followedId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId, followedId = params[0], params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the user is following the user to check
	exists, err := rt.db.FollowExists(otelctx, userId, followedId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if follow exists")
		http.Error(w, "Error checking if the follow exists.", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Not following the user.", http.StatusNotFound)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Follow{Follower: userId, Followed: followedId})
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the follow")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}
