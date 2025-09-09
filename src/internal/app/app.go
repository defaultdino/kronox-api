package app

import (
	"net/http"
	"time"

	"github.com/tumble-for-kronox/kronox-api/internal/clients/kronox"
)

type App struct {
	KronoxClient kronox.Client
}

func NewApp() (*App, error) {
	httpClient := &http.Client{
		Timeout: (30 * time.Second),
	}

	kronoxClient := kronox.NewClient(httpClient)

	return &App{
		KronoxClient: kronoxClient,
	}, nil
}
