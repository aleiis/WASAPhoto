package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleiis/WASAPhoto/service/api"
	"github.com/aleiis/WASAPhoto/service/database"

	"github.com/ardanlabs/conf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {

	// Load the configuration
	cfg, err := loadConfig()
	if errors.Is(err, conf.ErrHelpWanted) {
		return nil
	} else if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create the logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Infof("starting the application")
	logger.Infof("configuration loaded")

	// Create or open the database
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename+"?_foreign_keys=1")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("can't open the database: %w", err)
	}

	defer func() {
		logger.Debug("database stopping")
		dbconn.Close()
	}()

	db, err := database.New(dbconn)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return fmt.Errorf("creating AppDatabase: %w", err)
	}

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
	routerHandler := apirouter.Handler()

	/*
		router, err = registerWebUI(router)
		if err != nil {
			logger.WithError(err).Error("error registering web UI handler")
			return fmt.Errorf("registering web UI handler: %w", err)
		}
	*/

	// Apply CORS policy
	routerHandler = applyCORSHandler(routerHandler)

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           routerHandler,
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
			apiserver.Close()
		}
	}

	return nil
}
