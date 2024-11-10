/*
Webapi is the executable for the main web server.
It builds a web server around APIs from `service/api`.
Webapi connects to external resources needed (database) and starts two web servers: the API web server, and the debug.
Everything is served via the API web server, except debug variables (/debug/vars) and profiler infos (pprof).

Usage:

	webapi [flags]

Flags and configurations are handled automatically by the code in `load-configuration.go`.

Return values (exit codes):

	0
		The program ended successfully (no errors, stopped by signal)

	> 0
		The program ended due to an error

Note that this program will update the schema of the database to the latest version available (embedded in the
executable during the build).
*/
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aleiis/WASAPhoto/service/api"
	"github.com/aleiis/WASAPhoto/service/database"

	"github.com/ardanlabs/conf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// main is the program entry point. The only purpose of this function is to call run() and set the exit code if there is
// any error
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run executes the program. The body of this function should perform the following steps:
// * reads the configuration
// * creates and configure the logger
// * connects to any external resources (like databases, authenticators, etc.)
// * creates an instance of the service/api package
// * starts the principal web server (using the service/api.Router.Handler() for HTTP handlers)
// * waits for any termination event: SIGTERM signal (UNIX), non-recoverable server error, etc.
// * closes the principal web server
func run() error {

	// Load the configuration
	cfg, err := loadConfig()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return fmt.Errorf("loading config: %w", err)
	}

	// Create the logger
	logger := logrus.New()
	if cfg.LogFile != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755); err != nil {
			logger.WithError(err).Error("can't create the log directory")
			return fmt.Errorf("can't create the log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.WithError(err).Error("can't open the log file")
			return fmt.Errorf("can't open the log file: %w", err)
		}
		defer file.Close()
		logger.SetOutput(file)
	} else {
		logger.SetOutput(os.Stdout)
	}
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Infof("starting the application")
	logger.Infof("configuration loaded")

	// Check if all the folder structure exists
	if err := os.MkdirAll(filepath.Dir(cfg.DB.Filename), 0755); err != nil {
		logger.WithError(err).Error("can't create the folder structure for the database")
		return fmt.Errorf("can't create the folder structure for the database: %w", err)
	}

	// Create or open the database
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename+"?_foreign_keys=1")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("can't open the database: %w", err)
	}

	defer func() {
		logger.Debug("database stopping")
		_ = dbconn.Close()
	}()

	db, err := database.New(dbconn, cfg.DB.Filename)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}

	// Creates a context that will be used to cancel the logging goroutine
	ctx, cancelLogging := context.WithCancel(context.Background())
	defer cancelLogging() // Ensure the context is canceled when the function returns

	// Start the periodic logging of the database
	go database.StartPeriodicLogging(ctx, dbconn, time.Minute, cfg.DB.LogFile)

	// Start (main) API server
	logger.Info("initializing API server")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Create the API router
	apirouter, err := api.New(logger, db)
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return fmt.Errorf("registering web UI handler: %w", err)
	}

	// Apply CORS policy
	router = applyCORSHandler(router)

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	go func() {
		logger.Infof("API listening on %s", apiserver.Addr)
		serverErrors <- apiserver.ListenAndServe()
		logger.Infof("stopping API server")
	}()

	// Waiting for shutdown signal or POSIX signals
	select {
	case err := <-serverErrors:
		// Non-recoverable server error
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Infof("signal %v received, start shutdown", sig)

		// Asking API server to shut down and load shed.
		err := apirouter.Close()
		if err != nil {
			logger.WithError(err).Warning("graceful shutdown of apirouter error")
		}

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and load shed.
		err = apiserver.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Warning("error during graceful shutdown of HTTP server")
			_ = apiserver.Close()
		}
	}

	return nil
}
