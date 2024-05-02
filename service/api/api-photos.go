package api

import (
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
	"image"
	"net/http"
)

func (rt *_router) uploadPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

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
		if err == database.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			ctx.Logger.WithError(err).Error("can't upload the photo")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(201)
}

func (rt *_router) deletePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		photoId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Try to delete the photo
	if err := rt.db.DeletePhoto(userId, photoId); err != nil {
		if err == database.ErrPhotoNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			ctx.Logger.WithError(err).Error("can't delete the photo")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(200)
}
