package api

import (
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
)

const ContentTypeJPEG = "image/jpeg"
const ContentTypePNG = "image/png"

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
	if (contentTypeHeader != ContentTypeJPEG && contentTypeHeader != ContentTypePNG) ||
		(format != "jpeg" && format != "png") ||
		(contentTypeHeader == ContentTypeJPEG && format != "jpeg") ||
		(contentTypeHeader == ContentTypePNG && format != "png") {
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

func (rt *_router) getPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId, photoId = params[0], params[1]
	}

	// Check if the user requesting the profile is not banned by the user whose profile is being requested
	// The check is made using the Authorization header
	requesterId, err := getUserIdFromBearer(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	banCheck, err := rt.db.BanExists(userId, requesterId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user is banned")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if banCheck {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Try to get the path of the photo
	photo, err := rt.db.GetPhoto(userId, photoId)
	if err != nil {
		if err == database.ErrPhotoNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			ctx.Logger.WithError(err).Error("can't get the photo")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	photoPath := filepath.FromSlash(photo.Path)

	// Open the file
	fd, err := os.Open(photoPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't open the photo file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	// Get the extension of the photo to set the Content-Type header
	ext := filepath.Ext(photoPath)
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = ContentTypeJPEG
	case ".png":
		contentType = ContentTypePNG
	}

	// Configure the Content-Type header of the response
	w.Header().Set("Content-Type", contentType)

	// Copy the content of the file pointed by fd to the response writer
	if _, err := io.Copy(w, fd); err != nil {
		ctx.Logger.WithError(err).Error("can't copy the binary into the response writer")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
