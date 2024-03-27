# Build layer
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go test ./internal/...
# Static build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server ./cmd/main.go

# Run layer
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /app/server .
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]