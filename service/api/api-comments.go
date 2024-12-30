package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel/trace"
)

const maxCommentCount = 100000

type Comment struct {
	Owner     User          `json:"owner"`
	Photo     GlobalPhotoId `json:"photo"`
	CommentId int64         `json:"comment_id"`
	Content   string        `json:"content"`
}

type CommentRequest struct {
	OwnerId int64  `json:"owner_id"`
	Content string `json:"content"`
}

type PhotoComments struct {
	Comments []Comment `json:"comments"`
}

func (rt *_router) commentPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "commentPhotoHandler", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Subspan: Parameter validation
	var photoOwner, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId = params[0], params[1]
	}

	// Decode the user ID of comment owner and the content of the comment from the body of the request
	var commentRequest CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&commentRequest); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), commentRequest.OwnerId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
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

	// Check if the user ID of the user who commented the photo exists
	if exists, err := rt.db.UserExists(otelctx, commentRequest.OwnerId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "User commenting the photo does not exist.", http.StatusBadRequest)
		return
	}

	// Check if the comment length is between 1 and 128 bytes
	if !checkCommentContentFormat(commentRequest.Content) {
		http.Error(w, "Invalid comment content format.", http.StatusBadRequest)
		return
	}

	// Try to create the comment
	newCommentId, err := rt.db.CreateComment(otelctx, photoOwner, photoId, commentRequest.OwnerId, commentRequest.Content)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't create the comment")
		http.Error(w, "Error creating the comment.", http.StatusInternalServerError)
		return
	}

	var commentOwner User
	commentOwner.UserId = commentRequest.OwnerId
	commentOwner.Username, err = rt.db.GetUsername(otelctx, commentRequest.OwnerId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the username of the comment owner")
		http.Error(w, "Error getting the username of the comment owner.", http.StatusInternalServerError)
		return
	}

	newComment := Comment{
		Owner:     commentOwner,
		Photo:     GlobalPhotoId{photoOwner, photoId},
		CommentId: newCommentId,
		Content:   commentRequest.Content,
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newComment)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the response")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) uncommentPhotoHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "uncommentPhotoHandler", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get the parameters
	var photoOwner, photoId, commentId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId"), ps.ByName("commentId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId, commentId = params[0], params[1], params[2]
	}

	// Check if the comment exists
	if exists, err := rt.db.CommentExists(otelctx, photoOwner, photoId, commentId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the comment exists")
		http.Error(w, "Error checking if the comment exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Comment does not exists.", http.StatusNotFound)
		return
	}

	// Get the user ID of the comment owner
	commentOwner, err := rt.db.GetCommentOwner(otelctx, photoOwner, photoId, commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the comment owner")
		http.Error(w, "Error getting the comment owner.", http.StatusInternalServerError)
		return
	}

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), commentOwner) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Try to delete the comment
	err = rt.db.DeleteComment(otelctx, photoOwner, photoId, commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't delete the comment")
		http.Error(w, "Error deleting the comment.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (rt *_router) getCommentsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	otelctx, span := tracer.Start(r.Context(), "getCommentsHandler", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Get the parameters
	var photoOwner, photoId int64
	if params, err := checkIds(ps.ByName("userId"), ps.ByName("photoId")); err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	} else {
		photoOwner, photoId = params[0], params[1]
	}

	// Check if the user requesting the profile is not banned by the user whose profile is being requested
	// The check is made using the Authorization header
	banExists, err := checkBan(otelctx, rt.db, r.Header.Get("Authorization"), photoOwner)
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

	// Check if the photo exists
	if exists, err := rt.db.PhotoExists(otelctx, photoOwner, photoId); err != nil {
		ctx.Logger.WithError(err).Error("can't check if the photo exists")
		http.Error(w, "Error checking if the photo exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Photo not found.", http.StatusNotFound)
		return
	}

	// Try to get the comments of the photo
	comments, err := rt.db.GetPhotoComments(otelctx, photoOwner, photoId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the comments of the photo")
		http.Error(w, "Error getting the comments of the photo.", http.StatusInternalServerError)
		return
	}

	// Verify the number of comments
	if len(comments) > maxCommentCount {
		comments = comments[:maxCommentCount]
	}

	var photoComments PhotoComments
	photoComments.Comments = make([]Comment, len(comments))
	for i, comment := range comments {
		username, err := rt.db.GetUsername(otelctx, comment.CommentOwner)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get resolve the user of a comment")
			http.Error(w, "Error getting the username of a comment owner.", http.StatusInternalServerError)
			return
		}
		photoComments.Comments[i] = Comment{
			Owner:     User{UserId: comment.CommentOwner, Username: username},
			Photo:     GlobalPhotoId{photoOwner, photoId},
			CommentId: comment.CommentId,
			Content:   comment.Content,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(photoComments)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the comments")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}
