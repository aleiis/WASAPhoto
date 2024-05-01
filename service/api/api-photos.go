package api

import (
	"image"
	"net/http"
	"strconv"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) uploadPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	id := ps.ByName("userId")

	bearer := r.Header.Get("Authorization")

	if bearer != id {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user ID is a valid int64
	userId, err := strconv.ParseInt(id, 10, 64)
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

	// Read the img from the body of the request
	img, format, err := image.Decode(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the format of the image is jpeg or png
	contentTypeHeader := r.Header.Get("Content-Type")
	if (contentTypeHeader != "image/jpeg" && contentTypeHeader != "image/png") ||
		(format != "jpeg" && format != "png") ||
		(contentTypeHeader == "image/jpeg" && format != "jpeg") ||
		(contentTypeHeader == "image/png" && format != "png") {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Try to upload the photo
	err = rt.db.UploadPhoto(userId, img, format)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't upload the photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) deletePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	strUserId := ps.ByName("userId")
	strPhotoId := ps.ByName("photoId")

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

	// Check if the photo ID is a valid int64
	photoId, err := strconv.ParseInt(strPhotoId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the photo exists
	exists, err := rt.db.PhotoExists(userId, photoId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the photo exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Try to delete the photo
	err = rt.db.DeletePhoto(userId, photoId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't delete the photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
