# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build

WORKDIR /src

# Cache deps first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the repo
COPY . .

# Build the MCP server binary
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=${VERSION}" -o /out/mythic-mcp ./cmd/mythic-mcp


FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/mythic-mcp /usr/local/bin/mythic-mcp

# Default HTTP transport port
ENV MCP_HTTP_PORT=3333
ENV MCP_TRANSPORT=http
EXPOSE 3333

ENTRYPOINT ["/usr/local/bin/mythic-mcp"]
