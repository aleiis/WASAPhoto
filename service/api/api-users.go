package api

import (
	"encoding/json"
	"github.com/aleiis/WASAPhoto/service/database"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type PhotoIdentifier struct {
	OwnerId int64 `json:"ownerId"`
	PhotoId int64 `json:"photoId"`
}

type UserProfile struct {
	Username  string            `json:"username"`
	Photos    []PhotoIdentifier `json:"photos"`
	Uploads   int64             `json:"uploads"`
	Followers int64             `json:"followers"`
	Following int64             `json:"following"`
}

type StreamPhoto struct {
	Identifier PhotoIdentifier `json:"identifier"`
	DateTime   string          `json:"dateTime"`
	Likes      int64           `json:"likes"`
	Comments   int64           `json:"comments"`
}

func (rt *_router) setMyUserNameHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

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

	// Decode the new username from the body of the request
	var newUsername string
	if err := json.NewDecoder(r.Body).Decode(&newUsername); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username has the correct format
	if !checkUsername(newUsername) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to set the new username
	if err := rt.db.SetUsername(userId, newUsername); err != nil {
		if err == database.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else if err == database.ErrUsernameAlreadyExists {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			ctx.Logger.WithError(err).Error("can't set the username")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) getUserProfileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	if params, err := checkIds(ps.ByName("userId")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		userId = params[0]
	}

	var userProfile UserProfile

	// Get the username of the user
	if username, err := rt.db.GetUsername(userId); err != nil {
		if err == database.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			ctx.Logger.WithError(err).Error("can't get the username")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	} else {
		userProfile.Username = username
	}

	// Get the user photos
	userPhotos, err := rt.db.GetUserPhotos(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user photos")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userProfile.Photos = make([]PhotoIdentifier, len(userPhotos))
	for i, photo := range userPhotos {
		userProfile.Photos[i] = PhotoIdentifier{
			OwnerId: photo.UserId,
			PhotoId: photo.PhotoId,
		}
	}

	// Get the other user information
	userProfile.Uploads, userProfile.Followers, userProfile.Following, err = rt.db.GetUserProfileStats(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user profile stats")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Encode the user profile to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(userProfile)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user profile")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getMyStreamHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

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

	// Check if the user exists
	exists, err := rt.db.UserExists(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var stream []StreamPhoto

	// Get the stream of the user
	photos, err := rt.db.GetUserStream(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user stream")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stream = make([]StreamPhoto, len(photos))
	for i, photo := range photos {
		likes, comments, err := rt.db.GetPhotoStats(photo.UserId, photo.PhotoId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the photo stats")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stream[i] = StreamPhoto{
			Identifier: PhotoIdentifier{
				OwnerId: photo.UserId,
				PhotoId: photo.PhotoId,
			},
			DateTime: photo.Date,
			Likes:    likes,
			Comments: comments,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(stream)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user stream")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
