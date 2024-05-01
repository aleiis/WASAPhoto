package api

import (
	"strconv"
	"strings"
)

// checkBearer checks if the Authorization header is a valid Bearer token and
// if the bearer token matches the user token
func checkBearer(authHeader string, userId int64) bool {

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	authHeader = authHeader[7:]

	return authHeader == strconv.FormatInt(userId, 10)
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
