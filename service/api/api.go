package api

import (
	"errors"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Router is the package API interface representing an API handler builder
type RouterI interface {
	// Handler returns an HTTP handler for APIs provided in this package
	Handler() http.Handler

	// Close terminates any resource used in the package
	Close() error
}

type _router struct {
	router *httprouter.Router

	// baseLogger is a logger for non-requests contexts, like goroutines or background tasks not started by a request.
	// Use context logger if available (e.g., in requests) instead of this logger.
	baseLogger logrus.FieldLogger

	db database.AppDatabaseI
}

// New returns a new Router instance
func New(logger logrus.FieldLogger, db database.AppDatabaseI) (RouterI, error) {
	// Check if the configuration is correct
	if logger == nil {
		return nil, errors.New("logger is required")
	}
	if db == nil {
		return nil, errors.New("database is required")
	}

	// Create a new router where we will register HTTP endpoints. The server will pass requests to this router to be
	// handled.
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	return &_router{
		router:     router,
		baseLogger: logger,
		db:         db,
	}, nil
}
