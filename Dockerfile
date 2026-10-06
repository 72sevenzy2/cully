FROM golang:1.26.8-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /cully-mcp ./cmd/cully-mcp

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /cully-mcp /cully-mcp
COPY LICENSE NOTICE /licenses/
EXPOSE 8080
ENTRYPOINT ["/cully-mcp"]
