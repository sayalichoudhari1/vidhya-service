// Command vidhya-service starts the Vidhya School ERP HTTP API.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vidhya-service/src/app"
	"vidhya-service/src/logger"
	"vidhya-service/src/web"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.InitApp(ctx)
	if err != nil {
		logger.Log.Error("", "STARTUP_FAILED", "%s", err)
		os.Exit(1)
	}
	defer a.Cleanup()

	router := web.NewRouter(a)

	server := &http.Server{
		Addr:              a.Config.Server.Address,
		Handler:           router,
		ReadHeaderTimeout: a.Config.Server.ReadHeaderTimeout,
		ReadTimeout:       a.Config.Server.ReadTimeout,
		WriteTimeout:      a.Config.Server.WriteTimeout,
		IdleTimeout:       a.Config.Server.IdleTimeout,
	}

	go func() {
		logger.Log.Info("", "Vidhya Service listening on %s%s", a.Config.Server.Address, a.Config.Server.URLPrefix)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("", "SERVER_ERROR", "%s", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Log.Info("", "Shutdown signal received, draining in-flight requests...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("", "SHUTDOWN_ERROR", "%s", err)
	}

	logger.Log.Info("", "Vidhya Service stopped")
}
