package api

import (
	"encoding/json"
	"errors"
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

type GlobalPhotoId struct {
	Owner int64 `json:"owner"`
	Id    int64 `json:"id"`
}

type Photo struct {
	Owner    User   `json:"owner"`
	Id       int64  `json:"id"`
	Date     string `json:"date"`
	Likes    int64  `json:"likes"`
	Comments int64  `json:"comments"`
}

func (rt *_router) uploadPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	if params, err := checkIds(ps.ByName("userId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId = params[0]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the content type of the request is a valid image
	contentTypeHeader := r.Header.Get("Content-Type")
	if contentTypeHeader != ContentTypeJPEG && contentTypeHeader != ContentTypePNG {
		http.Error(w, "Invalid image format.", http.StatusBadRequest)
		return
	}

	// Read the img from the body of the request
	img, format, err := image.Decode(r.Body)
	if err != nil {
		http.Error(w, "Impossible to decode the image.", http.StatusBadRequest)
		return
	}

	// Check if the format of the image is jpeg or png
	if (format != "jpeg" && format != "png") ||
		(contentTypeHeader == ContentTypeJPEG && format != "jpeg") ||
		(contentTypeHeader == ContentTypePNG && format != "png") {
		http.Error(w, "The image format doesn't match the Content-Type header.", http.StatusBadRequest)
	}

	// Try to upload the photo
	err = rt.db.UploadPhoto(userId, img, format)
	switch {
	case errors.Is(err, database.ErrUserNotFound):
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	case err != nil:
		ctx.Logger.WithError(err).Error("can't upload the photo")
		http.Error(w, "Error saving the photo.", http.StatusInternalServerError)
		return
	}

	// Get the Global Photo ID of the uploaded photo
	lastUpload, err := rt.db.GetMostRecentPhoto(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the most recent photo")
		http.Error(w, "Error getting the global identifier of the uploaded photo.", http.StatusInternalServerError)
		return
	}

	// Write the Global Photo ID to the response
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(GlobalPhotoId{Owner: userId, Id: lastUpload.PhotoId})
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the global photo ID")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) deletePhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId = params[0]
		photoId = params[1]
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Try to delete the photo
	if err := rt.db.DeletePhoto(userId, photoId); err != nil {
		if errors.Is(err, database.ErrPhotoNotFound) {
			http.Error(w, "Photo not found.", http.StatusNotFound)
			return
		} else {
			ctx.Logger.WithError(err).Error("can't delete the photo")
			http.Error(w, "Error deleting the photo.", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(204)
}

func (rt *_router) getPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		userId, photoId = params[0], params[1]
	}

	// Check if the user requesting the photo is banned by the photo owner
	// The check is made using the Authorization header
	banExists, err := checkBan(rt.db, r.Header.Get("Authorization"), userId)
	switch {
	case errors.Is(err, ErrInvalidBearer):
		http.Error(w, "Invalid Bearer token.", http.StatusUnauthorized)
		return
	case err != nil:
		ctx.Logger.WithError(err).Error("can't check if the user is banned")
		http.Error(w, "Error checking if the user is banned.", http.StatusInternalServerError)
		return
	case banExists:
		http.Error(w, "User is banned by the owner of the photo.", http.StatusForbidden)
		return
	}

	// Try to get the path of the photo
	photoPath, err := rt.db.GetPhotoAbsolutePath(userId, photoId)
	switch {
	case errors.Is(err, database.ErrPhotoNotFound):
		http.Error(w, "Photo not found.", http.StatusNotFound)
		return
	case err != nil:
		ctx.Logger.WithError(err).Error("can't get the photo path")
		http.Error(w, "Error getting the photo path.", http.StatusInternalServerError)
		return
	}

	// Open the file
	fd, err := os.Open(photoPath)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't open the photo file")
		http.Error(w, "Error opening the photo file.", http.StatusInternalServerError)
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
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", contentType)

	// Copy the content of the file pointed by fd to the response writer
	if _, err := io.Copy(w, fd); err != nil {
		ctx.Logger.WithError(err).Error("can't copy the binary into the response writer")
		http.Error(w, "Error copying the photo into the response.", http.StatusInternalServerError)
		return
	}
}
