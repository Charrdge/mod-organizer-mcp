# Build (matches go.mod)
FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/mod-organizer-mcp ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mod-organizer-mcp /mod-organizer-mcp
EXPOSE 8080
ENTRYPOINT ["/mod-organizer-mcp"]
