package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type Comment struct {
	Owner   int64  `json:"owner"`
	Content string `json:"content"`
}

type ResolvedComment struct {
	User    string `json:"user"`
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
		fmt.Println(err)
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

func (rt *_router) getCommentsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var ownerId, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		ownerId, photoId = params[0], params[1]
	}

	// Check if the user requesting the profile is not banned by the user whose profile is being requested
	// The check is made using the Authorization header
	requesterId, err := getUserIdFromBearer(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	banCheck, err := rt.db.BanExists(ownerId, requesterId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user is banned")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if banCheck {
		w.WriteHeader(http.StatusForbidden)
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

	// Try to get the comments of the photo
	comments, err := rt.db.GetComments(ownerId, photoId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the comments of the photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resolvedComments []ResolvedComment
	resolvedComments = make([]ResolvedComment, len(comments))
	for i, comment := range comments {
		user, err := rt.db.GetUsername(comment.Owner)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get resolve the user of a comment")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resolvedComments[i] = ResolvedComment{
			User:    user,
			Content: comment.Content,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(resolvedComments)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the comments")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
