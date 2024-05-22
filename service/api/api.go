/*
Package api exposes the main API engine. All HTTP APIs are handled here - so-called "business logic" should be here, or
in a dedicated package (if that logic is complex enough).

To use this package, you should create a new instance with New() passing a valid logger and AppDatabase (see package
database). The resulting Router will have the Router.Handler() function that returns a handler that can be used in a
http.Server (or in other middlewares).

Example:

	// Create the API router
	apirouter, err := api.New(logger, app-db)
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("error creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	// ... other stuff here, like middleware chaining, etc.

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	apiserver.ListenAndServe()

See the `main.go` file inside the `cmd/webapi` for a full usage example.
*/

package api

import (
	"errors"
	"net/http"

	"github.com/aleiis/WASAPhoto/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// RouterI is the package API interface representing an API handler builder
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

// New returns a new Router instance which implements the RouterI interface. The router will be configured with the
// provided logger and database.
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
