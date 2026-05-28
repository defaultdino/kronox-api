package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/defaultdino/kronox-api/internal/app"
	"github.com/defaultdino/kronox-api/internal/server"
	"github.com/defaultdino/kronox-api/pkg/kronox"
	"github.com/gin-gonic/gin"
)

const (
	defaultPort     = "5055"
	shutdownTimeout = 30 * time.Second
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	if err := kronox.LoadSchools(); err != nil {
		a.Logger.Fatal().
			Err(err).
			Str("env_inline", kronox.EnvSchoolsJSON).
			Str("env_file", kronox.EnvSchoolsFile).
			Str("default_path", ".well-known/schools.json").
			Msg("Failed to load schools config")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	apiCfg := huma.DefaultConfig("Kronox API", "1.0.0")
	apiCfg.Info.Description = "Scrapes Kronox (https://www.kronox.se/) and returns JSON. See https://github.com/defaultdino/kronox-api."
	apiCfg.Info.License = &huma.License{Name: "Apache 2.0", URL: "https://www.apache.org/licenses/LICENSE-2.0.html"}
	api := humagin.New(r, apiCfg)

	server.New(a.Kronox, kronox.GetSchoolConfig()).Register(api)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	srv := &http.Server{Addr: ":" + port, Handler: r}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	a.Logger.Info().Str("port", port).Msg("Server started")

	<-quit
	a.Logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	a.Logger.Info().Msg("Server stopped")
}
