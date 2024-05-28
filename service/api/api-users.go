package api

import (
	"encoding/json"
	"errors"
	"github.com/aleiis/WASAPhoto/service/api/reqcontext"
	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
)

// maxStreamLength is the maximum number of photos that can be returned in a stream
const maxStreamLength = 100

// maxProfilePhotos is the maximum number of photos that can be returned in a profile
const maxProfilePhotos = 100000

type Username struct {
	Username string `json:"username"`
}

type User struct {
	UserId   int64  `json:"userId"`
	Username string `json:"username"`
}

type Profile struct {
	Owner     User    `json:"owner"`
	Photos    []Photo `json:"photos"`
	Uploads   int64   `json:"uploads"`
	Followers int64   `json:"followers"`
	Following int64   `json:"following"`
}

type Stream struct {
	Stream []Photo `json:"stream"`
}

func (rt *_router) setMyUserNameHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

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

	// Decode the new username from the body of the request
	var newUserResource User
	if err := json.NewDecoder(r.Body).Decode(&newUserResource); err != nil {
		http.Error(w, "Error decoding the request body.", http.StatusBadRequest)
		return
	}

	// Check if the user ID of the path and the user ID of the body match
	if userId != newUserResource.UserId {
		http.Error(w, "User ID from the path and the body don't match.", http.StatusBadRequest)
		return
	}

	// Check if the username has the correct format
	if !checkUsernameFormat(newUserResource.Username) {
		http.Error(w, "Invalid username format.", http.StatusBadRequest)
		return
	}

	// Try to set the new username
	if err := rt.db.SetUsername(userId, newUserResource.Username); err != nil {
		switch {
		case errors.Is(err, database.ErrUserNotFound):
			http.Error(w, "User not found.", http.StatusNotFound)
		case errors.Is(err, database.ErrUsernameAlreadyExists):
			http.Error(w, "Username already exists.", http.StatusConflict)
		default:
			ctx.Logger.WithError(err).Error("can't set the username")
			http.Error(w, "Error setting the username.", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(newUserResource)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the username")
		http.Error(w, "Error encoding the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getUserProfileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	params, err := checkIds(ps.ByName("userId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userId = params[0]

	// Check if the user requesting the profile is not banned by the user whose profile is being requested
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
		http.Error(w, "User is banned by the owner of the profile.", http.StatusForbidden)
		return
	}

	var userProfile Profile

	// Get the username of the user
	username, err := rt.db.GetUsername(userId)

	switch {
	case errors.Is(err, database.ErrUserNotFound):
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	case err != nil:
		ctx.Logger.WithError(err).Error("can't get the username")
		http.Error(w, "Error getting the username of the profile owner from the DB.", http.StatusInternalServerError)
		return
	}

	userProfile.Owner = User{
		UserId:   userId,
		Username: username,
	}

	// Get the user photos
	userPhotos, err := rt.db.GetUserPhotos(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user photos")
		http.Error(w, "Error getting the user photos.", http.StatusInternalServerError)
		return
	}

	if len(userPhotos) > maxProfilePhotos {
		userPhotos = userPhotos[:maxProfilePhotos]
	}

	userProfile.Photos = make([]Photo, len(userPhotos))
	for i, photo := range userPhotos {
		likes, comments, err := rt.db.GetPhotoStats(photo.UserId, photo.PhotoId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the photo stats")
			http.Error(w, "Error getting the stats of a photo.", http.StatusInternalServerError)
			return
		}
		userProfile.Photos[i] = Photo{
			Owner:    userProfile.Owner,
			Id:       photo.PhotoId,
			Date:     photo.Date,
			Likes:    likes,
			Comments: comments,
		}
	}

	// Get the other user information
	userProfile.Uploads, userProfile.Followers, userProfile.Following, err = rt.db.GetUserProfileStats(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user profile stats")
		http.Error(w, "Error getting the user profile stats.", http.StatusInternalServerError)
		return
	}

	// Encode the user profile to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(userProfile)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user profile")
		http.Error(w, "Error encoding the user profile in the response body.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getMyStreamHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	// Get the parameters
	var userId int64
	params, err := checkIds(ps.ByName("userId"))
	if err != nil {
		http.Error(w, "Missing or invalid parameters.", http.StatusBadRequest)
		return
	}
	userId = params[0]

	// Authorization check
	if !checkBearer(r.Header.Get("Authorization"), userId) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}

	// Check if the user exists
	exists, err := rt.db.UserExists(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't check if the user exists")
		http.Error(w, "Error checking if the user exists.", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	}

	var stream Stream

	// Get the photos of the user stream
	photos, err := rt.db.GetUserStream(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user stream")
		http.Error(w, "Error getting the user stream.", http.StatusInternalServerError)
		return
	}

	// Limit the number of photos in the stream
	if len(photos) > maxStreamLength {
		photos = photos[:maxStreamLength]
	}

	// Create the stream
	stream.Stream = make([]Photo, len(photos))
	for i, photo := range photos {
		likes, comments, err := rt.db.GetPhotoStats(photo.UserId, photo.PhotoId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the photo stats")
			http.Error(w, "Error getting the stats of a photo.", http.StatusInternalServerError)
			return
		}
		username, err := rt.db.GetUsername(photo.UserId)
		if err != nil {
			ctx.Logger.WithError(err).Error("can't get the username")
			http.Error(w, "Error getting the username of the owner of a photo.", http.StatusInternalServerError)
			return
		}
		stream.Stream[i] = Photo{
			Owner: User{UserId: userId,
				Username: username,
			},
			Id:       photo.PhotoId,
			Date:     photo.Date,
			Likes:    likes,
			Comments: comments,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(stream)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user stream")
		http.Error(w, "Error encoding the user stream.", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getUserByUsernameHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	var user User

	// Get the username from the URL
	parsedURL, err := url.Parse(r.RequestURI)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't parse the URL")
		http.Error(w, "Error parsing the URL.", http.StatusInternalServerError)
		return
	}

	queryParams := parsedURL.Query()
	user.Username = queryParams.Get("username")
	if len(user.Username) == 0 {
		http.Error(w, "Query parameter 'username' is missing.", http.StatusBadRequest)
		return
	}

	// Get the user id
	user.UserId, err = rt.db.GetUserId(user.Username)
	if errors.Is(err, database.ErrUserNotFound) {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	} else if err != nil {
		ctx.Logger.WithError(err).Error("can't get the user ID")
		http.Error(w, "Error getting the user ID.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		ctx.Logger.WithError(err).Error("can't encode the user")
		http.Error(w, "Error encoding the user.", http.StatusInternalServerError)
		return
	}
}
