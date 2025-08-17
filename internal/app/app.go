package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tumble-for-kronox/kronox-api/internal/clients/kronox"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	KronoxClient kronox.Client
	MongoDB      *mongo.Database
	Config       *Config
}

type Config struct {
	MongoDBURL  string        `json:"mongodb_url"`
	MongoDBName string        `json:"mongodb_name"`
	HTTPTimeout time.Duration `json:"http_timeout"`
	Port        string        `json:"port"`
	Environment string        `json:"environment"`
}

func NewApp() (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	httpClient := &http.Client{
		Timeout: config.HTTPTimeout,
	}

	kronoxClient := kronox.NewClient(httpClient)

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoDBURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	mongoDB := mongoClient.Database(config.MongoDBName)

	return &App{
		KronoxClient: kronoxClient,
		MongoDB:      mongoDB,
		Config:       config,
	}, nil
}

func (a *App) Close() error {
	if a.MongoDB != nil {
		return a.MongoDB.Client().Disconnect(context.Background())
	}
	return nil
}

// Loads a configuration file from the root directory,
// with an assumption that the config file is a JSON file in the
// format "env.<env>.json"
func loadConfig() (*Config, error) {
	env := flag.String("env", "dev", "Environment (dev, prod, test)")
	flag.Parse()

	configFile := fmt.Sprintf("env.%s.json", *env)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	config.Environment = *env

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second
	}

	return &config, nil
}
