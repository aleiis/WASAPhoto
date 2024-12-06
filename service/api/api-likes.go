package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Like struct {
	Liker int64         `json:"liker"`
	Photo GlobalPhotoId `json:"photo"`
}

func (rt *_router) likePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "likePhotoHandler")
	defer span.End()

	// Get the parameters
	var photoOwner, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId = params[0], params[1]
	}

	// Decode the user ID of the user who liked the photo from the body of the request
	var like Like
	if err := json.NewDecoder(r.Body).Decode(&like); err != nil {
		http.Error(w, "Error decoding the request body.", http.StatusBadRequest)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), like.Liker) {
		http.Error(w, "Unathorized.", http.StatusUnauthorized)
		return
	}

	// Check if the photo exists
	if exists, err := rt.db.PhotoExists(otelctx, photoOwner, photoId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the photo exists")
		http.Error(w, "Error checking if the photo exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Photo not found.", http.StatusNotFound)
		return
	}

	// Check if the user ID of the user who liked the photo exists
	if exists, err := rt.db.UserExists(otelctx, like.Liker); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "User liking the photo does not exist.", http.StatusBadRequest)
		return
	}

	// Check if the path parameters match the body parameters
	if photoOwner != like.Photo.OwnerId || photoId != like.Photo.PhotoId {
		http.Error(w, "Photo ID mismatch. The photo ID in the URL must be the same as the one in the request body.", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked the photo
	exists, err := rt.db.LikeExists(otelctx, like.Photo.OwnerId, like.Photo.PhotoId, like.Liker)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the like exists")
		http.Error(w, "Error checking if the like exists.", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "Already liked the photo.", http.StatusBadRequest)
		return
	}

	// Try to like the photo
	err = rt.db.CreateLike(otelctx, like.Photo.OwnerId, like.Photo.PhotoId, like.Liker)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't like the photo")
		http.Error(w, "Error liking the photo.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(like)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the response")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
	}
}

func (rt *_router) unlikePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "unlikePhotoHandler")
	defer span.End()

	// Get the parameters
	var photoOwner, photoId, likerId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId"), ps.ByName("likerId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId, likerId = params[0], params[1], params[2]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), likerId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the like exists
	exists, err := rt.db.LikeExists(otelctx, photoOwner, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the like exists")
		http.Error(w, "Error checking if the like exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Like not found.", http.StatusNotFound)
		return
	}

	// Try to unlike the photo
	err = rt.db.DeleteLike(otelctx, photoOwner, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't unlike the photo")
		http.Error(w, "Error unliking the photo.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (rt *_router) checkLikeStatusHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "checkLikeStatusHandler")
	defer span.End()

	// Get the parameters
	var photoOwner, photoId, likerId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId"), ps.ByName("likerId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId, likerId = params[0], params[1], params[2]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), likerId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the user has liked the photo
	exists, err := rt.db.LikeExists(otelctx, photoOwner, photoId, likerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check like status")
		http.Error(w, "Error checking if the like exists.", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Like does not exists.", http.StatusNotFound)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Like{Liker: likerId, Photo: GlobalPhotoId{OwnerId: photoOwner, PhotoId: photoId}})
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the response")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
	}
}
