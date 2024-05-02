package api

import (
	"encoding/json"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Comment struct {
	Owner   int64  `json:"owner"`
	Content string `json:"content"`
}

func (rt *_router) commentPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var ownerId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		ownerId, photoId = params[0], params[1]
	}

	// Decode the user ID of comment owner and the content of the comment from the body of the request
	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), comment.Owner) {
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

	// Check if the user ID of the user who commented the photo exists
	if exists, err := rt.db.UserExists(comment.Owner); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the comment length is between 1 and 128 bytes
	if len(comment.Content) == 0 || len(comment.Content) > 128 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to create the comment
	err := rt.db.CreateComment(ownerId, photoId, comment.Owner, comment.Content)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't create the comment")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(201)
}

func (rt *_router) uncommentPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var ownerId, photoId, commentId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId"), ps.ByName("commentId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		ownerId, photoId, commentId = params[0], params[1], params[2]
	}

	// Check if the comment exists
	if exists, err := rt.db.CommentExists(ownerId, photoId, commentId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the comment exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get the user ID of the comment owner
	commentOwner, err := rt.db.GetCommentOwner(ownerId, photoId, commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the comment owner")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), commentOwner) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Try to delete the comment
	err = rt.db.DeleteComment(ownerId, photoId, commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't delete the comment")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}
