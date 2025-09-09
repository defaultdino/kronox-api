FROM golang:alpine AS builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kronox-api cmd/kronox-api/main.go

FROM scratch
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/kronox-api .
COPY --from=builder /app/schools.toml ./
EXPOSE 5055
CMD ["/app/kronox-api"]