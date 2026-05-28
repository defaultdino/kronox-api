# kronox-api

A Go library and HTTP service that scrapes the [Kronox](https://www.kronox.se/) university schedule system and returns JSON.

Kronox itself exposes data as XML (`SchemaXML.jsp`) and HTML (`ajax_sokResurser.jsp`).

## Two ways to use it

### As a Go library

```go
import "github.com/defaultdino/kronox-api/pkg/kronox"

c := kronox.New()

events, err := c.GetEvents(ctx, kronox.EventsRequest{
    BaseURL:     "https://kronox.hkr.se",
    SchoolCode:  "hkr",
    ScheduleIDs: []string{"p.TBIT2+2026+35+100+NML+en"},
})

programmes, err := c.SearchProgrammes(ctx, kronox.ProgrammesRequest{
    BaseURL: "https://kronox.hkr.se",
    Query:   "data",
})
```

Pass `WithHTTPClient` or `WithUserAgent` to customize transport behavior:

```go
c := kronox.New(
    kronox.WithHTTPClient(myClient),
    kronox.WithUserAgent("my-app/1.0"),
)
```

Errors from Kronox surface as `*kronox.APIError`:

```go
if apiErr, ok := kronox.IsAPIError(err); ok {
    log.Printf("kronox returned %d", apiErr.StatusCode)
}
```

You can also call the parsers directly if you have a Kronox response from elsewhere:

```go
events, _ := kronox.ParseScheduleXML("hkr", []string{"p.X"}, xmlString)
programmes, _ := kronox.ParseProgrammes(htmlString)
```

### As an HTTP service

Build and run the binary:

```sh
go build -o kronox-api ./cmd/kronox-api
./kronox-api
```

Or via Docker:

```sh
docker build -t kronox-api .
docker run -p 5055:5055 \
  -e KRONOX_SCHOOLS_JSON='{"schools":{"hkr":{"id":"hkr","name":"HKR","domain":"hkr.se","urls":["https://kronox.hkr.se"]}}}' \
  kronox-api
```

Endpoints (under `/api/v1`):

- `GET /api/v1/schedule/events?school=hkr&url_index=0&schedule_ids=...[&start_date=YYYY-MM-DD]`
- `GET /api/v1/programme/search?school=hkr&url_index=0&q=...`

Utility endpoints:

- `GET /health` — liveness check
- `GET /schools` — list of allowed school codes
- `GET /schools/{school}/urls` — list of URLs for a school

The service is auth-less by design — if you need API keys, rate limiting, or caching, put a layer in front of it.

### API documentation

The HTTP service exposes its API description as **OpenAPI 3.1**, generated at runtime from the typed handler signatures via [huma](https://github.com/danielgtaylor/huma).

| Path | What you get |
|---|---|
| `GET /docs` | Browser UI ([Stoplight Elements](https://github.com/stoplightio/elements)) |
| `GET /openapi.json` | Full spec, JSON |
| `GET /openapi.yaml` | Full spec, YAML |
| `GET /openapi-3.0.json` / `.yaml` | OpenAPI 3.0 downgrade for older tooling |
| `GET /schemas/{type}.json` | Per-type JSON Schema; this is what the `$schema` URL in response bodies points at |

Success responses include a `$schema` field pointing back at the relevant schema URL. Clients that want strict validation can follow it.

## Configuration

### Schools

The service needs to know which schools it's allowed to scrape. Provide config in one of three ways (in precedence order):

1. **`KRONOX_SCHOOLS_JSON`** — inline JSON, useful for Kubernetes ConfigMaps.
2. **`KRONOX_SCHOOLS_FILE`** — path to a JSON file.
3. **Default** — `./.well-known/schools.json` relative to the process cwd.

The JSON shape:

```json
{
  "schools": {
    "hkr": {
      "id": "hkr",
      "name": "Högskolan Kristianstad",
      "domain": "hkr.se",
      "urls": ["https://kronox.hkr.se", "https://schema.hkr.se"],
      "logoUrl": ""
    }
  }
}
```

`urls` is an ordered list of Kronox base URLs for the school. HTTP clients pass `url_index` to pick which one to hit, so you can serve multiple subdomains per school (e.g. `kronox.*` and `schema.*`). Library consumers pass the chosen URL directly via `BaseURL`.

The service refuses to start if no source resolves.

If you're embedding the library in a larger Go program, you can also load the config yourself and inject it:

```go
import "github.com/defaultdino/kronox-api/pkg/kronox"

kronox.SetSchools(kronox.SchoolsConfig{
    Schools: map[string]kronox.School{
        "hkr": {Id: "hkr", URLs: []string{"https://kronox.hkr.se"}},
    },
})
```

`SetSchools` bypasses the env-var/file pipeline. Use it when your program already owns the config (e.g., reading from a database, a discovery service, or a different file layout).

### Other env vars

| Var | Default | Notes |
|---|---|---|
| `PORT` | `5055` | HTTP listen port |
| `LOG_LEVEL` | `info` | zerolog level: `trace`, `debug`, `info`, `warn`, `error` |

### Library knobs (not exposed via the HTTP layer)

`EventsRequest`:

- `StartDate` (`*time.Time`) — defaults to today
- `Intervals` — months ahead, defaults to `6`
- `Language` — defaults to `EN`

`ProgrammesRequest`:

- `StartDate` / `EndDate` (`*time.Time`)
- `IntervalType` — defaults to `m` (months)
- `Intervals` — defaults to `6`

## Limitations

- Kronox returns XML for schedules and HTML for programmes. Parsing depends on the upstream rendering — schema or markup changes will require parser updates.
- The library doesn't handle URL failover or caching. Those policies belong to the consumer.
- Programme search results return up to whatever Kronox returns; there's no built-in pagination.

## License

Apache 2.0 — see [LICENSE](LICENSE).
