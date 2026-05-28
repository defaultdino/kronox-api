package app

import (
	"net/http"
	"os"
	"time"

	"github.com/defaultdino/kronox-api/pkg/kronox"
	"github.com/rs/zerolog"
)

type App struct {
	Kronox *kronox.Client
	Logger *zerolog.Logger
}

func NewApp() (*App, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()

	return &App{
		Kronox: kronox.New(kronox.WithHTTPClient(httpClient)),
		Logger: &logger,
	}, nil
}
