package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.POST("/session", rt.wrap(rt.doLoginHandler))
	rt.router.GET("/users/", rt.wrap(rt.getUserByUsernameHandler))
	rt.router.PUT("/users/:userId", rt.wrap(rt.setMyUserNameHandler))
	rt.router.GET("/users/:userId/profile", rt.wrap(rt.getUserProfileHandler))
	rt.router.GET("/users/:userId/stream", rt.wrap(rt.getMyStreamHandler))
	rt.router.POST("/users/:userId/photos/", rt.wrap(rt.uploadPhotoHandler))
	rt.router.DELETE("/users/:userId/photos/:photoId", rt.wrap(rt.deletePhotoHandler))
	rt.router.GET("/users/:userId/photos/:photoId/bin", rt.wrap(rt.getPhotoHandler))
	rt.router.POST("/users/:userId/follows/", rt.wrap(rt.followUserHandler))
	rt.router.DELETE("/users/:userId/follows/:followedId", rt.wrap(rt.unfollowUserHandler))
	rt.router.GET("/users/:userId/follows/:followedId", rt.wrap(rt.checkFollowHandler))
	rt.router.POST("/users/:userId/bans/", rt.wrap(rt.banUserHandler))
	rt.router.DELETE("/users/:userId/bans/:bannedId", rt.wrap(rt.unbanUserHandler))
	rt.router.GET("/users/:userId/bans/:bannedId", rt.wrap(rt.checkBanHandler))
	rt.router.POST("/users/:userId/photos/:photoId/likes/", rt.wrap(rt.likePhotoHandler))
	rt.router.DELETE("/users/:userId/photos/:photoId/likes/:likerId", rt.wrap(rt.unlikePhotoHandler))
	rt.router.GET("/users/:userId/photos/:photoId/likes/:likerId", rt.wrap(rt.checkLikeStatusHandler))
	rt.router.POST("/users/:userId/photos/:photoId/comments/", rt.wrap(rt.commentPhotoHandler))
	rt.router.GET("/users/:userId/photos/:photoId/comments/", rt.wrap(rt.getCommentsHandler))
	rt.router.DELETE("/users/:userId/photos/:photoId/comments/:commentId", rt.wrap(rt.uncommentPhotoHandler))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
