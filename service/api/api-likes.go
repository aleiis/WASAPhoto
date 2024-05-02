package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) likePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var ownerId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		ownerId, photoId = params[0], params[1]
	}

	// Decode the user ID of the user who liked the photo from the body of the request
	var likerId int64
	if err := json.NewDecoder(r.Body).Decode(&likerId); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), likerId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the photo exists
	if exists, err := rt.db.PhotoExists(ownerId, photoId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the photo exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if the user ID of the user who liked the photo exists
	if exists, err := rt.db.UserExists(likerId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check if the user has already liked the photo
	exists, err := rt.db.LikeExists(ownerId, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the like exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to like the photo
	err = rt.db.CreateLike(ownerId, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't like the photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) unlikePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var ownerId, photoId, likerId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId"), ps.ByName("likerId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		ownerId, photoId, likerId = params[0], params[1], params[2]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), likerId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user has already liked the photo
	exists, err := rt.db.LikeExists(ownerId, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the like exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to unlike the photo
	err = rt.db.DeleteLike(ownerId, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unlike the photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
