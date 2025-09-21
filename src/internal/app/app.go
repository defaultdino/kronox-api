package app

import (
	"net/http"
	"time"

	"github.com/tumble-for-kronox/kronox-api/internal/clients/kronox"
)

type App struct {
}

func NewApp() (*App, error) {
	return &App{}, nil
}

func (a *App) NewKronoxClient() kronox.Client {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return kronox.NewClient(httpClient)
}
