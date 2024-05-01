package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aleiis/WASAPhoto/service/database"

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

	id := ps.ByName("userId")

	bearer := r.Header.Get("Authorization")

	if bearer != id {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Decode the new username from the body of the request
	var newUsername string
	if err := json.NewDecoder(r.Body).Decode(&newUsername); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the new username contains at least 3 characters and no more than 16
	if len(newUsername) < 3 || len(newUsername) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username contains only alphanumeric characters
	for _, c := range newUsername {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Check if the user ID is a valid int64
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
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

	id := ps.ByName("userId")

	// Check if the user ID is a valid int64
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var userProfile UserProfile

	// Get the username of the user
	userProfile.Username, err = rt.db.GetUsername(userId)
	if err != nil {
		if err == database.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			ctx.Logger.WithError(err).Error("can't get the username")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Get the user photos
	userPhotos, err := rt.db.GetUserPhotos(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user photos")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		userProfile.Photos = make([]PhotoIdentifier, len(userPhotos))
		for i, photo := range userPhotos {
			userProfile.Photos[i] = PhotoIdentifier{
				OwnerId: photo.UserId,
				PhotoId: photo.PhotoId,
			}
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

	strUserId := ps.ByName("userId")

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

	var stream []StreamPhoto

	// Get the stream of the user
	photos, err := rt.db.GetUserStream(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user stream")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
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
