package api

import (
	"strconv"
	"strings"
)

// getBearer returns the Bearer token of a given user
func getBearer(userId int64) (string, error) {
	return strconv.FormatInt(userId, 10), nil
}

// checkBearer checks if the Authorization header is a valid Bearer token and
// if the bearer token matches the user token
func checkBearer(authHeader string, userId int64) bool {

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	authHeader = authHeader[7:]

	userToken, err := getBearer(userId)
	if err != nil {
		return false
	}

	return authHeader == userToken
}

// checkIds checks if the given string IDs are valid and returns them
// with the correct type, otherwise it returns an error
func checkIds(strIds ...string) ([]int64, error) {
	var ids []int64

	for _, strId := range strIds {
		id, err := strconv.ParseInt(strId, 10, 64)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// checkUsername checks if the given username is valid
func checkUsername(username string) bool {
	if len(username) < 3 || len(username) > 16 {
		return false
	}

	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}

	return true
}
