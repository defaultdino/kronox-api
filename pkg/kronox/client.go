package kronox

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/defaultdino/kronox-api/pkg/models"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

type Client struct {
	httpClient *http.Client
	userAgent  string
}

type Option func(*Client)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  defaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type APIError struct {
	StatusCode int
	Status     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kronox returned %s", e.Status)
}

type EventsRequest struct {
	BaseURL     string
	SchoolCode  string
	ScheduleIDs []string
	StartDate   *time.Time
	Intervals   int
	Language    string
}

func (c *Client) GetEvents(ctx context.Context, req EventsRequest) ([]*models.Event, error) {
	if req.Intervals == 0 {
		req.Intervals = 6
	}
	if req.Language == "" {
		req.Language = "EN"
	}
	startDatum := "idag"
	if req.StartDate != nil {
		startDatum = req.StartDate.Format("2006-01-02")
	}
	endpoint := strings.TrimSuffix(req.BaseURL, "/") + "/setup/jsp/SchemaXML.jsp"
	params := map[string]string{
		"startDatum":     startDatum,
		"intervallTyp":   "m",
		"intervallAntal": strconv.Itoa(req.Intervals),
		"sprak":          req.Language,
		"sokMedAND":      "false",
		"forklaringar":   "true",
		"resurser":       strings.Join(req.ScheduleIDs, ","),
	}
	body, err := c.fetch(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}
	return ParseScheduleXML(req.SchoolCode, req.ScheduleIDs, body)
}

type ProgrammesRequest struct {
	BaseURL      string
	Query        string
	StartDate    *time.Time
	EndDate      *time.Time
	IntervalType string
	Intervals    int
}

func (c *Client) SearchProgrammes(ctx context.Context, req ProgrammesRequest) ([]*models.Programme, error) {
	if req.IntervalType == "" {
		req.IntervalType = "m"
	}
	if req.Intervals == 0 {
		req.Intervals = 6
	}
	startDatum := "idag"
	if req.StartDate != nil {
		startDatum = req.StartDate.Format("2006-01-02")
	}
	slutDatum := ""
	if req.EndDate != nil {
		slutDatum = req.EndDate.Format("2006-01-02")
	}
	endpoint := strings.TrimSuffix(req.BaseURL, "/") + "/ajax/ajax_sokResurser.jsp"
	params := map[string]string{
		"sokord":         req.Query,
		"startDatum":     startDatum,
		"slutDatum":      slutDatum,
		"intervallTyp":   req.IntervalType,
		"intervallAntal": strconv.Itoa(req.Intervals),
	}
	body, err := c.fetch(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}
	return ParseProgrammes(body)
}

func IsAPIError(err error) (*APIError, bool) {
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}

func (c *Client) fetch(ctx context.Context, endpoint string, params map[string]string) (string, error) {
	return doRequest(ctx, c.httpClient, c.userAgent, endpoint, params)
}
