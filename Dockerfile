FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/kronox-api ./cmd/kronox-api

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/kronox-api .
EXPOSE 5055
CMD ["/app/kronox-api"]
