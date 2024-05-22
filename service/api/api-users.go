package api

import (
	"encoding/json"
	"fmt"
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
)

type User struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
}

type Username struct {
	Username string `json:"username"`
}

type PhotoIdentifier struct {
	OwnerId int64 `json:"ownerId"`
	PhotoId int64 `json:"photoId"`
}

type ProfilePhoto struct {
	PhotoId  int64  `json:"photoId"`
	DateTime string `json:"dateTime"`
	Likes    int64  `json:"likes"`
	Comments int64  `json:"comments"`
}

type UserProfile struct {
	Username  string         `json:"username"`
	Photos    []ProfilePhoto `json:"photos"`
	Uploads   int64          `json:"uploads"`
	Followers int64          `json:"followers"`
	Following int64          `json:"following"`
}

type StreamPhoto struct {
	Identifier PhotoIdentifier `json:"identifier"`
	User       string          `json:"user"`
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
	var newUsername Username
	if err := json.NewDecoder(r.Body).Decode(&newUsername); err != nil {
		fmt.Println(r.ContentLength)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username has the correct format
	if !checkUsername(newUsername.Username) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to set the new username
	if err := rt.db.SetUsername(userId, newUsername.Username); err != nil {
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

	userProfile.Photos = make([]ProfilePhoto, len(userPhotos))
	for i, photo := range userPhotos {
		likes, comments, err := rt.db.GetPhotoStats(userId, photo.PhotoId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the photo stats")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userProfile.Photos[i] = ProfilePhoto{
			PhotoId:  photo.PhotoId,
			DateTime: photo.Date,
			Likes:    likes,
			Comments: comments,
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
		user, err := rt.db.GetUsername(photo.UserId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the username")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stream[i] = StreamPhoto{
			Identifier: PhotoIdentifier{
				OwnerId: photo.UserId,
				PhotoId: photo.PhotoId,
			},
			User:     user,
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

func (rt *_router) getUserByUsernameHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	var user User

	parsedURL, err := url.Parse(r.RequestURI)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't parse the URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queryParams := parsedURL.Query()
	user.Username = queryParams.Get("username")
	if len(user.Username) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the user id
	user.UserId, err = rt.db.GetUserId(user.Username)
	if err != nil {
		if err == database.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			ctx.Logger.WithError(err).Error("can't find the user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
