package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/defaultdino/kronox-api/pkg/kronox"
	"github.com/defaultdino/kronox-api/pkg/models"
)

type Server struct {
	client  *kronox.Client
	schools kronox.SchoolsConfig
}

func New(client *kronox.Client, schools kronox.SchoolsConfig) *Server {
	return &Server{client: client, schools: schools}
}

func (s *Server) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-schedule-events",
		Method:      http.MethodGet,
		Path:        "/api/v1/schedule/events",
		Summary:     "Get schedule events",
		Description: "Fetches schedule events from Kronox for one or more schedule IDs.",
		Tags:        []string{"Schedules"},
	}, s.GetScheduleEvents)

	huma.Register(api, huma.Operation{
		OperationID: "search-programmes",
		Method:      http.MethodGet,
		Path:        "/api/v1/programme/search",
		Summary:     "Search programmes",
		Description: "Free-text search over Kronox programmes for a given school.",
		Tags:        []string{"Programmes"},
	}, s.SearchProgrammes)

	huma.Register(api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Liveness check",
		Tags:        []string{"Utility"},
	}, s.Health)

	huma.Register(api, huma.Operation{
		OperationID: "list-schools",
		Method:      http.MethodGet,
		Path:        "/schools",
		Summary:     "List configured school codes",
		Tags:        []string{"Utility"},
	}, s.ListSchools)

	huma.Register(api, huma.Operation{
		OperationID: "school-urls",
		Method:      http.MethodGet,
		Path:        "/schools/{school}/urls",
		Summary:     "List Kronox base URLs for a school",
		Tags:        []string{"Utility"},
	}, s.SchoolURLs)
}

type GetEventsInput struct {
	School      string `query:"school" required:"true" example:"hkr"`
	URLIndex    int    `query:"url_index" required:"true" minimum:"0" example:"0"`
	ScheduleIDs string `query:"schedule_ids" required:"true" doc:"Comma-separated list of schedule IDs"`
	StartDate   string `query:"start_date" doc:"Start date in YYYY-MM-DD format; defaults to today"`
}

type GetEventsOutput struct {
	Body struct {
		Events []*models.Event `json:"events"`
	}
}

func (s *Server) GetScheduleEvents(ctx context.Context, in *GetEventsInput) (*GetEventsOutput, error) {
	baseURL, err := s.schoolURL(in.School, in.URLIndex)
	if err != nil {
		return nil, err
	}

	ids := strings.Split(in.ScheduleIDs, ",")

	var startDate *time.Time
	if in.StartDate != "" {
		t, perr := time.Parse("2006-01-02", in.StartDate)
		if perr != nil {
			return nil, huma.Error400BadRequest("invalid start_date format, use YYYY-MM-DD")
		}
		startDate = &t
	}

	events, err := s.client.GetEvents(ctx, kronox.EventsRequest{
		BaseURL:     baseURL,
		SchoolCode:  in.School,
		ScheduleIDs: ids,
		StartDate:   startDate,
	})
	if err != nil {
		if _, ok := kronox.IsAPIError(err); ok {
			return nil, huma.Error502BadGateway("upstream Kronox returned an error")
		}
		return nil, huma.Error500InternalServerError("failed to fetch schedule events")
	}

	out := &GetEventsOutput{}
	if events == nil {
		out.Body.Events = []*models.Event{}
	} else {
		out.Body.Events = events
	}
	return out, nil
}

type SearchProgrammesInput struct {
	School   string `query:"school" required:"true" example:"hkr"`
	URLIndex int    `query:"url_index" required:"true" minimum:"0" example:"0"`
	Query    string `query:"q" required:"true" doc:"Free-text search query"`
}

type SearchProgrammesOutput struct {
	Body struct {
		Programmes []*models.Programme `json:"programmes"`
	}
}

func (s *Server) SearchProgrammes(ctx context.Context, in *SearchProgrammesInput) (*SearchProgrammesOutput, error) {
	baseURL, err := s.schoolURL(in.School, in.URLIndex)
	if err != nil {
		return nil, err
	}

	programmes, err := s.client.SearchProgrammes(ctx, kronox.ProgrammesRequest{
		BaseURL: baseURL,
		Query:   in.Query,
	})
	if err != nil {
		if _, ok := kronox.IsAPIError(err); ok {
			return nil, huma.Error502BadGateway("upstream Kronox returned an error")
		}
		return nil, huma.Error500InternalServerError("failed to fetch programmes")
	}

	out := &SearchProgrammesOutput{}
	if programmes == nil {
		out.Body.Programmes = []*models.Programme{}
	} else {
		out.Body.Programmes = programmes
	}
	return out, nil
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"healthy"`
	}
}

func (s *Server) Health(ctx context.Context, _ *struct{}) (*HealthOutput, error) {
	out := &HealthOutput{}
	out.Body.Status = "healthy"
	return out, nil
}

type ListSchoolsOutput struct {
	Body struct {
		AllowedSchools []string `json:"allowed_schools"`
	}
}

func (s *Server) ListSchools(ctx context.Context, _ *struct{}) (*ListSchoolsOutput, error) {
	out := &ListSchoolsOutput{}
	out.Body.AllowedSchools = make([]string, 0, len(s.schools.Schools))
	for code := range s.schools.Schools {
		out.Body.AllowedSchools = append(out.Body.AllowedSchools, code)
	}
	return out, nil
}

type SchoolURLsInput struct {
	School string `path:"school"`
}

type SchoolURLEntry struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

type SchoolURLsOutput struct {
	Body struct {
		School string           `json:"school"`
		URLs   []SchoolURLEntry `json:"urls"`
	}
}

func (s *Server) SchoolURLs(ctx context.Context, in *SchoolURLsInput) (*SchoolURLsOutput, error) {
	school, ok := s.schools.Schools[in.School]
	if !ok {
		return nil, huma.Error404NotFound("school not found")
	}
	out := &SchoolURLsOutput{}
	out.Body.School = in.School
	out.Body.URLs = make([]SchoolURLEntry, len(school.URLs))
	for i, u := range school.URLs {
		out.Body.URLs[i] = SchoolURLEntry{Index: i, URL: u}
	}
	return out, nil
}

func (s *Server) schoolURL(code string, urlIndex int) (string, error) {
	school, ok := s.schools.Schools[code]
	if !ok {
		return "", huma.Error400BadRequest(fmt.Sprintf("unknown school %q", code))
	}
	if urlIndex < 0 || urlIndex >= len(school.URLs) {
		return "", huma.Error400BadRequest(fmt.Sprintf("url_index %d out of range (school has %d URLs)", urlIndex, len(school.URLs)))
	}
	return school.URLs[urlIndex], nil
}
